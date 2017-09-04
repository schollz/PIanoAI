package ai

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/schollz/gobrain"
	"github.com/schollz/pianoai/music"
	log "github.com/sirupsen/logrus"
)

// Learn is for calculating the matricies for the transition
// probabilities
func (ai *AI) Learn2(notes music.Notes) (err error) {
	logger := log.WithFields(log.Fields{
		"function": "AI.Learn2",
	})
	if ai.IsLearning {
		return errors.New("Already learning")
	}
	ai.toggleLearning(true)
	defer ai.toggleLearning(false)

	// Analyze the notes
	logger.Info("Analyzing notes")
	ai.notes = ai.Analyze(notes)
	if len(ai.notes) < 10 {
		return errors.New("Need more 30 notes")
	}

	patterns := [][][]float64{}
	for i, note := range ai.notes {
		ai.notes[i] = append(note, 0)
	}
	for i, note := range ai.notes {
		if i == 0 {
			continue
		}
		previousNote := convertIntsToFloats(ai.notes[i-1])
		currentNote := convertIntsToFloats(note)
		pattern := [][]float64{
			previousNote, currentNote,
		}
		patterns = append(patterns, pattern)
	}
	// instantiate the Feed Forward
	ai.ff = &gobrain.FeedForward{}

	// initialize the Neural Network;
	// the networks structure will contain:
	// 2 inputs, 2 hidden nodes and 1 output.
	logger.Debug("Initializing neural net...")
	ai.ff.Init(5*32, 30, 5*32)

	// train the network using the XOR patterns
	// the training will run for 1000 epochs
	// the learning rate is set to 0.6 and the momentum factor to 0.4
	// use true in the last parameter to receive reports about the learning error
	logger.Debug("Training neural net...")
	ai.ff.Train(patterns, 10000, 0.1, 0.3, true)
	logger.Debug("Finished training.")
	ai.HasLearned = true
	return
}

// Lick2 generates a sequence of chords using the Markov
// probabilities. Must run Learn2() beforehand.
func (ai *AI) Lick2(startBeat int) (lick *music.Music, err error) {
	logger := log.WithFields(log.Fields{
		"function": "AI.Lick2",
	})
	if !ai.HasLearned || ai.IsLearning {
		err = errors.New("Learning must be finished")
		return
	}

	// // Generate lick from the neural network
	logger.Debug("Generating lick from neural net")
	notes := [][]int{}
	note := ai.notes[rand.Intn(len(ai.notes))] // Pick a random note
	for {
		noteInput := convertIntsToFloats(note)
		noteOutput := ai.ff.Update(noteInput)
		note = convertFloatToInts(noteOutput)
		fmt.Println(note)
		notes = append(notes, note)
		if len(notes) > ai.MaximumLickLength {
			if len(notes) < ai.MinimumLickLength {
				continue
			}
			break
		}
	}

	// Convert the notes to a music
	lick = ConvertNotes(notes, startBeat)
	return
}

func convertFloatToInts(f []float64) []int {
	m := make([]int, len(f)/32)
	curVal := 0
	total := 0
	for i, val := range f {
		x := 31 - math.Mod(float64(i), 32)
		if int(val+0.5) == 1 { // round
			total += int(math.Pow(2, float64(x)))
		}
		if x == 0 {
			m[curVal] = total
			total = 0
			curVal++
		}
	}
	return m
}
func convertIntsToFloats(ns []int) []float64 {
	m := make([]float64, 32*len(ns))
	for i, n := range ns {
		for j, newN := range convertIntToFloats(n) {
			m[j+i*32] = newN
		}
	}
	return m
}

func convertIntToFloats(i int) []float64 {
	m := make([]float64, 32)
	mi := 31
	for _, c := range reverse(strconv.FormatInt(int64(i), 2)) {
		if c == '1' {
			m[mi] = 1
		} else {
			m[mi] = 0
		}
		mi--
	}
	return m
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
