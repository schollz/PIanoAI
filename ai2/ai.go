package ai2

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/schollz/rpiai-piano/music"
	log "github.com/sirupsen/logrus"
	hashids "github.com/speps/go-hashids"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type AI struct {
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

	// LinkLength is how many links should be used
	LinkLength int

	// WindowSize is how many total notes to include
	WindowSize int

	hasher           *hashids.HashIDData
	links            map[string]string
	notes            music.Note
	chords           map[string][]Chord
	chordArray       []Chord
	chordStringArray []string
}

type Chord struct {
	Pitches  []int
	Velocity int
	Duration int
	Lag      int
}

func New() (ai *AI) {
	ai = new(AI)
	ai.HighPassFilter = 60
	ai.MinimumLickLength = 2
	ai.MaximumLickLength = 30
	ai.hasher = hashids.NewData()
	ai.hasher.Salt = "piano"
	ai.hasher.MinLength = 8
	ai.LinkLength = 2
	ai.WindowSize = 5
	return ai
}

func (ai *AI) toggleLearning(l bool) {
	ai.IsLearning = l
}

func (ai *AI) encode(ints []int) string {
	h, _ := hashids.NewWithData(ai.hasher)
	e, _ := h.Encode(ints)
	return e
}

func (ai *AI) decode(s string) []int {
	h, _ := hashids.NewWithData(ai.hasher)
	return h.Decode(s)
}

func (ai *AI) Learn(music *music.Music) (analyzedNotes [][]int) {
	logger := log.WithFields(log.Fields{
		"function": "AI.Analyze",
	})
	logger.Debug("Analyzing...")
	// initialize the links and the chords
	ai.links = make(map[string]string)
	ai.chords = make(map[string][]Chord)

	// sort the beats
	beats := make([]int, len(music.Notes))
	beatI := 0
	for beat := range music.Notes {
		beats[beatI] = beat
		beatI++
	}
	sort.Ints(beats)

	ai.chordArray = make([]Chord, len(beats))
	ai.chordStringArray = make([]string, len(beats))
	chordArrayI := 0
	for _, beat1 := range beats {
		chord := Chord{
			Pitches: []int{},
		}
		duration := 0
		lag := 0
		velocity := 0

		for note1 := range music.Notes[beat1] {
			if !music.Notes[beat1][note1].On || note1 < ai.HighPassFilter {
				continue
			}
			chord.Pitches = append(chord.Pitches, note1)
			if velocity == 0 {
				velocity = music.Notes[beat1][note1].Velocity
			}
			if duration > 0 && lag > 0 {
				continue
			}
			// determine duration and lag
			for _, beat2 := range beats {
				if beat2 < beat1 {
					continue
				}
				if lag == 0 {
					lag = beat2 - beat1
				}
				for note2 := range music.Notes[beat2] {
					if note2 == note1 && !music.Notes[beat2][note2].On {
						duration = beat2 - beat1
						break
					}
				}
			}
		}
		if len(chord.Pitches) == 0 {
			continue
		}
		chord.Velocity = velocity
		chord.Duration = duration
		chord.Lag = lag
		chordString := ai.encode(chord.Pitches)
		if _, ok := ai.chords[chordString]; !ok {
			ai.chords[chordString] = []Chord{}
		}
		ai.chords[chordString] = append(ai.chords[chordString], chord)
		ai.chordStringArray[chordArrayI] = chordString
		ai.chordArray[chordArrayI] = chord
		chordArrayI++
	}
	ai.chordArray = ai.chordArray[:chordArrayI]
	ai.chordStringArray = ai.chordStringArray[:chordArrayI]
	logger.Debugf("...analyzed %d chords", len(ai.chordArray))
	ai.HasLearned = true
	return
}

// Lick generates a sequence of chords using the Markov
// probabilities. Must run Learn() beforehand.
func (ai *AI) Lick(startBeat int) (lick *music.Music, err error) {
	if !ai.HasLearned || ai.IsLearning {
		err = errors.New("Learning must be finished")
		return
	}
	lick = music.New()

	// expanded to allow it to wrap
	chordStringArray := append(ai.chordStringArray[(len(ai.chordStringArray)-ai.WindowSize):], ai.chordStringArray...)
	chordStringArray = append(chordStringArray, ai.chordStringArray[:ai.WindowSize]...)

	start := rand.Intn(len(ai.chordArray))
	song := []int{}

	for {
		// ending criteria
		if len(song) > 10 {
			break
		}
		// add the chord indicies to the song
		for i := 0; i < ai.WindowSize; i++ {
			startI := start + i
			if startI >= len(ai.chordStringArray) {
				startI = 0
				start = -1 * i
			}
			song = append(song, startI)
		}

		// find a new start sequence that is the same as the end sequence of the current song
		sequenceToFind := make([]string, ai.LinkLength)
		for i := 0; i < ai.LinkLength; i++ {
			sequenceToFind[i] = ai.chordStringArray[song[len(song)-(ai.LinkLength-i)]]
		}
		// find the starts of that sequence
		fmt.Println(sequenceToFind)
		candidateStarts := []int{}
		for i := range chordStringArray {
			if i < ai.WindowSize || i > len(chordStringArray)-ai.WindowSize {
				continue
			}
			foundMatch := true
			for j := 0; j < ai.LinkLength; j++ {
				if chordStringArray[i+j] != sequenceToFind[j] {
					foundMatch = false
					break
				}
			}
			if foundMatch {
				candidateStarts = append(candidateStarts, i)
			}
		}

		// pick a new start
		start = candidateStarts[rand.Intn(len(candidateStarts))] - ai.WindowSize + ai.LinkLength
	}

	// fmt.Println(song)
	// for i, s := range ai.chordStringArray {
	// 	fmt.Println(i, s)
	// }

	// make them into a song
	firstBeat := startBeat
	for _, index := range song {
		for _, pitch := range ai.chordArray[index].Pitches {
			lick.AddNote(music.Note{
				On:       true,
				Pitch:    pitch,
				Velocity: ai.chordArray[index].Velocity,
				Beat:     firstBeat,
			})
			lick.AddNote(music.Note{
				On:       false,
				Pitch:    pitch,
				Velocity: ai.chordArray[index].Velocity,
				Beat:     firstBeat + ai.chordArray[index].Duration,
			})
		}
		firstBeat += ai.chordArray[index].Lag
	}
	return
}
