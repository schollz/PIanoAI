package ai

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/schollz/gobrain"
	"github.com/schollz/rpiai-piano/music"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

// MarkovAI is an implementation of an AI that aims to
// improvise in realtime. In this implementation, the current
// history of real playing is used to generate transition
// probabilities which are used to reconstruct new licks.
type AI struct {
	// BeatsBetweenLicks specifies the amount of space
	// between each lick before adding an "end"
	BeatsBetweenLicks int

	// HighPassFilter only uses notes above a certain level
	HighPassFilter int

	// MinimumLickLength is the minimum number of notes for a lick
	MinimumLickLength int

	// MaximumLickLength is the maximum number of notes for a lick
	MaximumLickLength int

	// keep track of whether it is learning,
	// so learning can be done asynchronously
	IsLearning bool
	HasLearned bool

	// transition matrix for probabilities
	// I is the INDEX of the property in question,
	// where {0,1,2,3} -> {Pitch,Velocity,Duration,Lag}
	// A and B are prior probabilities. A and B are some
	// current or previous property in the sequence.
	// C is the value of the property in quesiton.
	// D is the probability of obtaining C given A,B.
	// I.e. F = P(C|A,B) for property 'I'
	//           I        A      B       C   D
	matrices map[int]map[int]map[int]map[int]int

	// Basic coupling, each state only depends on previous
	// Coupling index determines what it is coupled to (P/V/D/L)
	// Coupling code:
	//  0 signals to ignore
	// -1 couples that index to previous state at coupling index
	// 1 couples that index to current state at coupling index
	// e.g. []][]int{{-1, 1, 0, 0}, ... }  couples the CURRENT pitch
	// to the PREVIOUS pitch AND the CURRENT velocity.
	// Note: Only the first two couplings will be used.
	coupling [][]int

	// List of all the notes and their properties
	// {0,1,2,3} -> {Pitch,Velocity,Duration,Lag}
	notes [][]int

	// Order to process notes in
	stateOrdering []int

	// neural network
	ff *gobrain.FeedForward
}

func New() (m *AI) {
	m = new(AI)
	// matrices initialized to handle all four indices
	m.matrices = make(map[int]map[int]map[int]map[int]int)
	for i := 0; i <= 3; i++ {
		m.matrices[i] = make(map[int]map[int]map[int]int)
	}

	m.coupling = [][]int{{-1, 0, 0, 0}, {0, -1, 0, 0}, {0, 0, -1, 0}, {0, 0, 0, -1}}
	m.notes = [][]int{}
	m.stateOrdering = []int{0, 1, 2, 3}
	m.BeatsBetweenLicks = 16 * 64
	m.HighPassFilter = 60
	m.MinimumLickLength = 2
	m.MaximumLickLength = 30
	return m
}

// Couple will take an index and a coupling and
// attach to the matrix.
// For example, to couple current Velocity to
// previous Pitch and previous Velocity, you would
// use Couple(1,[]int{-1,-1,0,0}),
// where {0,1,2,3} -> {Pitch,Velocity,Duration,Lag}.
func (m *AI) Couple(index int, coupling []int) {

	m.coupling[index] = coupling
}

func (m *AI) toggleLearning(l bool) {

	m.IsLearning = l
}

func (m *AI) Analyze(notes music.Notes) (analyzedNotes [][]int) {

	analyzedNotes = [][]int{}
	sort.Sort(notes)
	// Find a note that turns on
	for i, note1 := range notes {
		if !note1.On {
			continue
		}
		//              Pitch         Velocity      Duration Lag
		values := []int{note1.Pitch, note1.Velocity, 10000, 10000}
		// Loop to find the Duration and the Lag
		for j, note2 := range notes {
			// Only consider notes after the current
			if j <= i {
				continue
			}
			// Find when the current note turns off to get the Duration
			if !note2.On && note1.Pitch == note2.Pitch && values[2] == 10000 {
				values[2] = note2.Beat - note1.Beat
			}
			// Find when next note turns on to get the Lag
			if note2.On && values[3] == 10000 {
				values[3] = note2.Beat - note1.Beat
			}
			// If the values are filled, then append and move on
			if values[2] != 10000 && values[3] != 10000 && values[0] >= m.HighPassFilter {
				analyzedNotes = append(analyzedNotes, values)
				break
			}
		}
	}
	return
}

func (m *AI) addToMatrices(i, a, b, c, d int) {

	if _, ok := m.matrices[i][a]; !ok {
		m.matrices[i][a] = make(map[int]map[int]int)
	}
	if _, ok := m.matrices[i][a][b]; !ok {
		m.matrices[i][a][b] = make(map[int]int)
	}
	if _, ok := m.matrices[i][a][b][c]; !ok {
		m.matrices[i][a][b][c] = 0
	}
	m.matrices[i][a][b][c] += d
}

// Learn is for calculating the matricies for the transition
// probabilities
func (m *AI) Learn(notes music.Notes) (err error) {
	logger := log.WithFields(log.Fields{
		"function": "AI.Learn",
	})
	if m.IsLearning {
		return errors.New("Already learning")
	}
	m.toggleLearning(true)
	defer m.toggleLearning(false)

	// Analyze the notes
	logger.Info("Analyzing notes")
	m.notes = m.Analyze(notes)
	if len(m.notes) < 10 {
		return errors.New("Need more 30 notes")
	}

	// Determine transition frequencies for the corresponding couplings, and then normalize
	logger.Info("Determine transition frequencies")
	for _, i := range m.stateOrdering { // i is the index of the property
		prevValue := []int{-1, -1, -1, -1}
		curValue := []int{-1, -1, -1, -1}
		a := -1
		b := -1
		for noteNum, note := range m.notes {
			// logger.Debugf("note: %+v", note)
			curValue = note
			a = -1
			b = -1
			insufficientInfo := false // coupling must be done in the correct order (left to user)
			for index, place := range m.coupling[i] {
				logger.Debugf("curValue: %+v, prevValue: %+v", curValue, prevValue)
				logger.Debugf("index: %+v, place: %+v", index, place)
				if place == 0 {
					// ignore this coupling
					continue
				} else if place == -1 {
					if prevValue[index] == -1 {
						insufficientInfo = true
						break
					}
					if a == -1 {
						a = prevValue[index]
					} else if b == -1 {
						b = prevValue[index]
					}
				} else if place == 1 {
					if curValue[index] == -1 {
						insufficientInfo = true
						break
					}
					if a == -1 {
						a = curValue[index]
					} else if b == -1 {
						b = curValue[index]
					}
				}
			}
			if insufficientInfo {
				logger.Warnf("Insufficient info for a: %+v,b: %+v", a, b)
			} else {
				m.addToMatrices(i, a, b, note[i], 1)
			}
			if noteNum == 0 || (noteNum > 0 && prevValue[3] > m.BeatsBetweenLicks) {
				// this starts a lick
				m.addToMatrices(i, -200, -200, curValue[i], 1)
				if noteNum > 0 && i == 0 {
					m.addToMatrices(i, a, b, -404, 1)
				}
			} else if len(m.notes)-1 == noteNum && i == 0 {
				// this ends a lick
				m.addToMatrices(i, a, b, -404, 1)
			}
			prevValue = curValue
		}
		// -404 signals end

	}

	// Normalize the transitions
	logger.Debug("Normalize transitions")

	for i := range m.matrices {
		for a := range m.matrices[i] {
			for b := range m.matrices[i][a] {

				// Determine probability
				total := 0
				for _, d := range m.matrices[i][a][b] {
					total += d
				}
				for c, d := range m.matrices[i][a][b] {
					m.matrices[i][a][b][c] = (d * 10000) / total // generates a number between 0 - 10000
				}

				// determine cumulative sum
				intKeys := make([]int, len(m.matrices[i][a][b]))
				index := 0
				for c := range m.matrices[i][a][b] {
					intKeys[index] = c
					index++
				}
				sort.Ints(intKeys)
				prevValue := 0
				for _, c := range intKeys {
					m.matrices[i][a][b][c] = prevValue + m.matrices[i][a][b][c]
					prevValue = m.matrices[i][a][b][c]
				}
			}
		}
	}
	m.HasLearned = true

	return
}

// Lick generates a sequence of chords using the Markov
// probabilities. Must run Learn() beforehand.
func (m *AI) Lick(startBeat int) (lick *music.Music, err error) {

	if !m.HasLearned || m.IsLearning {
		err = errors.New("Learning must be finished")
		return
	}

	// // Generate lick from the transition probabilities
	// // by looping through properties in the order specified.
	notes := [][]int{}
	note := []int{-1, -1, -1, -1}
	for {
		note = m.GenerateNote(note)
		if note[0] == -404 || len(notes) > m.MaximumLickLength {
			if len(notes) < m.MinimumLickLength {
				continue
			}
			break
		}
		notes = append(notes, note)
	}
	fmt.Println(notes)

	// Convert the notes to a music
	lick = ConvertNotes(notes, startBeat)

	return
}

func ConvertNotes(notes [][]int, startBeat int) (song *music.Music) {
	song = music.New()
	curBeat := startBeat
	for _, note := range notes {
		song.AddNote(music.Note{
			On:       true,
			Pitch:    note[0],
			Velocity: note[1],
			Beat:     curBeat,
		})
		song.AddNote(music.Note{
			On:       false,
			Pitch:    note[0],
			Velocity: 0,
			Beat:     curBeat + note[2],
		})
		curBeat += note[3]
	}
	return song
}

func (m *AI) GenerateNote(prevValue []int) (curValue []int) {

	curValue = []int{-1, -1, -1, -1}
	for _, i := range m.stateOrdering {
		a := -1
		b := -1
		if prevValue[i] != -1 {
			// First pick the first note
			for index, place := range m.coupling[i] {
				if place == 0 {
					// ignore this coupling
					continue
				} else if place == -1 {
					if a == -1 {
						a = prevValue[index]
					} else if b == -1 {
						b = prevValue[index]
					}
				} else if place == 1 {
					if a == -1 {
						a = curValue[index]
					} else if b == -1 {
						b = curValue[index]
					}
				}
			}
		} else {
			a = -200
			b = -200
		}
		if len(m.matrices[i][a][b]) == 0 {
			a = -200
			b = -200
		}
		curValue[i] = pickRandom(m.matrices[i][a][b])
	}
	return
}

func pickRandom(m map[int]int) (picked int) {
	r := rand.Intn(10000)
	for _, p := range rankByProb(m) {
		picked = p.Key
		if r <= p.Value {
			return
		}
	}
	return
}

func rankByProb(stateFrequencies map[int]int) PairList {
	pl := make(PairList, len(stateFrequencies))
	i := 0
	for k, v := range stateFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   int
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value > p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
