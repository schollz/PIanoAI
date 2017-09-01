package ai

import (
	"errors"
	"sort"

	"github.com/schollz/pianoai/music"
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
	// keep track of whether it is learning,
	// so learning can be done asynchronously
	IsLearning bool
	HasLearned bool

	// transition matrix for probabilities
	// I -> A -> B -> C
	// I is the INDEX of the property in question,
	// where {0,1,2,3} -> {Pitch,Velocity,Duration,Lag}
	// A and B are VALUES of the current/previous properties.
	// C is a probability, which will later be
	// transformed to the cumulative
	// probability (for Markov transitions).
	matrices map[int]map[int]map[int]int

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
}

func New() (m *AI) {
	m = new(AI)
	// matrices initialized to handle all four indices
	m.matrices = make(map[int]map[int]map[int]int)
	for i := 0; i <= 3; i++ {
		m.matrices[i] = make(map[int]map[int]int)
	}

	m.coupling = [][]int{{-1, 0, 0, 0}, {0, -1, 0, 0}, {0, 0, -1, 0}, {0, 0, 0, -1}}
	m.notes = [][]int{}
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

func (m *AI) Analyze(notes music.Notes) {
	sort.Sort(notes)
	for i, note1 := range notes {
		if !note1.On {
			continue
		}
		// Find a note that turns on
		values := []int{note1.Pitch, note1.Velocity, 10000, 10000}
		// Loop to find the Duration and the Lag
		for j, note2 := range notes {
			// Only consider notes after the current
			if j < i {
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
			if values[2] != 10000 && values[3] != 10000 {
				m.notes = append(m.notes, values)
				break
			}
		}
	}

}

// Learn is for calculating the matricies for the transition
// probabilities
func (m *AI) Learn(notes music.Notes) (err error) {

	m.toggleLearning(true)
	defer m.toggleLearning(false)

	// Analyze the notes
	m.Analyze(notes)

	// TODO: Determine transition frequencies
	// for the corresponding couplings, and then normalize
	stateOrdering := []int{0, 1, 2, 3} // determine Pitch, Velocity, Duration, and Lag in that order
	for _, state := range stateOrdering {
		prevValue := []int{-1, -1, -1, -1}
		curValue := []int{-1, -1, -1, -1}
		for _, note := range m.notes {
			curValue = note
			A := -1
			B := -1
			insufficientInfo := false // coupling must be done in the correct order (left to user)
			for index, place := range m.coupling[state] {
				if place == 0 {
					// ignore this coupling
					continue
				} else if place == -1 {
					if prevValue[index] == -1 {
						insufficientInfo = true
						break
					}
					if A == -1 {
						A = prevValue[index]
					} else if B == -1 {
						B = prevValue[index]
					}
				} else if place == 1 {
					if curValue[index] == -1 {
						insufficientInfo = true
						break
					}
					if A == -1 {
						A = curValue[index]
					} else if B == -1 {
						B = curValue[index]
					}
				}
			}
			if insufficientInfo {
				continue
			}
			if _, ok := m.matrices[state][A]; !ok {
				m.matrices[state][A] = make(map[int]int)
			}
			if _, ok := m.matrices[state][A][B]; !ok {
				m.matrices[state][A][B] = 0
			}
			m.matrices[state][A][B]++
			prevValue = curValue
		}
	}

	// Normalize the transitions
	for s := range m.matrices {
		for a := range m.matrices[s] {
			// Determine probability
			total := 0
			for b := range m.matrices[s][a] {
				total += m.matrices[s][a][b]
			}
			for b := range m.matrices[s][a] {
				m.matrices[s][a][b] = (m.matrices[s][a][b] * 10000) / total // generates a number between 0 - 10000
			}

			// determine cumulative sum
			intKeys := make([]int, len(m.matrices[s][a]))
			i := 0
			for b := range m.matrices[s][a] {
				intKeys[i] = b
			}
			sort.Ints(intKeys)
			prevValue := 0
			for _, b := range intKeys {
				m.matrices[s][a][b] = prevValue + m.matrices[s][a][b]
				prevValue = m.matrices[s][a][b]
			}

		}
	}

	m.HasLearned = true
	return
}

// Lick generates a sequence of chords using the Markov
// probabilities. Must run Learn() beforehand.
func (m *AI) Lick() (lick *music.Music, err error) {
	lick = music.New()
	if !m.HasLearned || m.IsLearning {
		err = errors.New("Learning must be finished")
		return
	}

	// TODO: Generate lick from the transition probabilities

	return
}
