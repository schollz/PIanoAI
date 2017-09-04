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
	WindowSizeMin, WindowSizeMax int

	hasher           *hashids.HashIDData
	links            map[string]string
	notes            music.Note
	chords           map[string][]Chord
	chordArray       []Chord
	chordStringArray []string

	Jazzy          bool
	Stacatto       bool
	DisallowChords bool

	MaxChordDistance int
}

type Chord struct {
	Pitches  []int
	Velocity int
	Duration int
	Lag      int
}

func New() (ai *AI) {
	ai = new(AI)
	ai.HighPassFilter = 65
	ai.MinimumLickLength = 2
	ai.MaximumLickLength = 30
	ai.hasher = hashids.NewData()
	ai.hasher.Salt = "piano"
	ai.hasher.MinLength = 8
	ai.LinkLength = 3
	ai.WindowSizeMin = 30
	ai.WindowSizeMax = 50
	ai.Jazzy = true
	ai.DisallowChords = true
	ai.MaxChordDistance = 6 // TODO: THIS SHOULD DEPEND ON BPM
	ai.Stacatto = true
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

func (ai *AI) Learn(mus *music.Music) (err error) {
	logger := log.WithFields(log.Fields{
		"function": "AI.Analyze",
	})
	logger.Debug("Analyzing...")
	// initialize the links and the chords
	ai.links = make(map[string]string)
	ai.chords = make(map[string][]Chord)

	// sanitize music (merge chords)
	skipBeats := make(map[int]bool)
	for beat1 := range mus.Notes {
		for beat2 := range mus.Notes {
			if beat2 < beat1 || beat2-beat1 > ai.MaxChordDistance {
				continue
			}
			for note := range mus.Notes[beat2] {
				if _, ok := mus.Notes[beat1][note]; !ok {
					mus.Notes[beat1][note] = music.Note{
						On:       mus.Notes[beat2][note].On,
						Pitch:    mus.Notes[beat2][note].Pitch,
						Velocity: mus.Notes[beat2][note].Velocity,
						Beat:     mus.Notes[beat2][note].Beat,
					}
					skipBeats[beat2] = true
				}
			}
		}
	}

	// sort the beats
	beats := make([]int, len(mus.Notes))
	beatI := 0
	for beat := range mus.Notes {
		if _, ok := skipBeats[beat]; ok {
			continue
		}
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

		for note1 := range mus.Notes[beat1] {
			if !mus.Notes[beat1][note1].On || note1 < ai.HighPassFilter || mus.Notes[beat1][note1].Velocity < 70 {
				continue
			}
			chord.Pitches = append(chord.Pitches, note1)
			if velocity == 0 {
				velocity = mus.Notes[beat1][note1].Velocity
			}
			if duration > 0 && lag > 0 {
				continue
			}
			// determine duration and lag
			for _, beat2 := range beats {
				if beat2 <= beat1 {
					continue
				}
				for note2 := range mus.Notes[beat2] {
					if lag == 0 && mus.Notes[beat2][note2].On {
						lag = beat2 - beat1
					}
					if duration == 0 && note2 == note1 && !mus.Notes[beat2][note2].On {
						duration = beat2 - beat1
					}
				}
			}
		}
		if len(chord.Pitches) == 0 {
			continue
		}
		chord.Velocity = velocity
		chord.Duration = duration
		if lag > 64*4 {
			lag = 64 * 4
		}
		fmt.Println(duration, lag)
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
	logger := log.WithFields(log.Fields{
		"function": "AI.Lick",
	})

	if !ai.HasLearned || ai.IsLearning {
		err = errors.New("Learning must be finished")
		return
	}
	lick = music.New()

	start := rand.Intn(len(ai.chordArray))
	song := []int{}

	for {
		// expanded to allow it to wrap
		windowSize := ai.WindowSizeMin - rand.Intn(ai.WindowSizeMax-ai.WindowSizeMin)
		chordStringArray := append(ai.chordStringArray[(len(ai.chordStringArray)-windowSize-1):], ai.chordStringArray...)
		chordStringArray = append(chordStringArray, ai.chordStringArray[:windowSize+1]...)

		// add the chord indicies to the song
		for i := 0; i < windowSize; i++ {
			startI := start + i
			if startI >= len(ai.chordStringArray) {
				startI = 0
				start = -1 * i
			}
			song = append(song, startI)
		}

		// ending criteria
		lickLength := 0
		for _, index := range song {
			lickLength += ai.chordArray[index].Lag
		}
		if lickLength > 64*16 {
			break
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
			if i < windowSize || i > len(chordStringArray)-windowSize {
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
		if len(candidateStarts) == 0 {
			start += windowSize
		} else {
			start = candidateStarts[rand.Intn(len(candidateStarts))] - windowSize + ai.LinkLength + 1
		}
	}

	// fmt.Println(song)
	// for i, s := range ai.chordStringArray {
	// 	fmt.Println(i, s)
	// }

	// make them into a song
	firstBeat := startBeat
	quantizer := 8
	for _, index := range song {
		extraDuration := 0
		stacatto := 0
		if ai.Stacatto {
			stacatto = 2
		}
		if ai.Jazzy {
			if rand.Intn(20) == 1 {
				extraDuration += 64 * (1 + rand.Intn(4))
			}
		}

		for _, pitch := range ai.chordArray[index].Pitches {
			logger.Debugf("Adding note %d @ %d with lag %d", pitch, (firstBeat)/quantizer*quantizer, ai.chordArray[index].Lag)
			onNote := music.Note{
				On:       true,
				Pitch:    pitch,
				Velocity: ai.chordArray[index].Velocity,
				Beat:     (firstBeat) / quantizer * quantizer,
			}
			offNote := music.Note{
				On:       false,
				Pitch:    pitch,
				Velocity: 0,
				Beat:     (firstBeat+ai.chordArray[index].Duration)/quantizer*quantizer + extraDuration,
			}
			if offNote.Beat-onNote.Beat > 16 {
				offNote.Beat -= stacatto
			}
			lick.AddNote(onNote)
			lick.AddNote(offNote)
			if ai.DisallowChords {
				logger.Debug("Disallowing chord, breaking")
				break
			}
		}
		firstBeat += (ai.chordArray[index].Lag)/quantizer*quantizer + extraDuration + stacatto
		if ai.Jazzy {
			if rand.Intn(10) == 1 {
				firstBeat += 64
			}
		}
	}
	return
}
