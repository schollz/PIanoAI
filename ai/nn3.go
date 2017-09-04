package ai

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/schollz/gobrain"
	"github.com/schollz/pianoai/music"
	log "github.com/sirupsen/logrus"
)

// Learn is for calculating the matricies for the transition
// probabilities
func (ai *AI) Learn4(notes music.Notes) (err error) {
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

	emptyPiano := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	previousPiano := emptyPiano
	currentPiano := emptyPiano

	patterns := [][][]float64{}
	for i, note := range ai.notes {
		currentPiano[note[0]] = 1
		if ai.notes[i][3] == 0 || i == 0 {
			continue
		}
		pattern := [][]float64{
			previousPiano, currentPiano,
		}
		fmt.Println("----")
		fmt.Println(previousPiano)
		fmt.Println(currentPiano)
		patterns = append(patterns, pattern)
		previousPiano = currentPiano
		currentPiano = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}

	// instantiate the Feed Forward
	ai.ff2[0] = &gobrain.FeedForward{}
	logger.Debug("Initializing neural net...")
	ai.ff2[0].Init(1*len(emptyPiano), 10, 1*len(emptyPiano))
	logger.Debug("Training neural net...")
	ai.ff2[0].Train(patterns, 2000, 0.1, 0.4, true)
	logger.Debug("Finished training.")

	ai.HasLearned = true
	return
}

// Lick2 generates a sequence of chords using the Markov
// probabilities. Must run Learn2() beforehand.
func (ai *AI) Lick4(startBeat int) (lick *music.Music, err error) {
	logger := log.WithFields(log.Fields{
		"function": "AI.Lick2",
	})
	if !ai.HasLearned || ai.IsLearning {
		err = errors.New("Learning must be finished")
		return
	}

	// // Generate lick from the neural network
	logger.Debug("Generating lick from neural net")
	emptyPiano := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	note := ai.notes[rand.Intn(len(ai.notes))] // Pick a random note
	currentPiano := emptyPiano
	currentPiano[note[0]] = 1
	j := 0
	notes := [][]int{}
	for {
		currentPiano = ai.ff2[j].Update(currentPiano)
		newNotes := getNotesFromPiano(currentPiano)
		if len(newNotes) == 0 {
			note = ai.notes[rand.Intn(len(ai.notes))] // Pick a random note
			currentPiano = emptyPiano
			currentPiano[note[0]] = 1
			continue
		}
		fmt.Println(newNotes)
		notes = append(notes, newNotes...)

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

func getNotesFromPiano(p []float64) [][]int {
	notes := [][]int{}
	for i, val := range p {
		if val > 0.5 {
			notes = append(notes, []int{i, 50, 7, 0})
		}
	}
	if len(notes) > 0 {
		notes[len(notes)-1][3] = 8 // set the time till next note
	}
	return notes
}

func sumSlice(n []float64) float64 {
	total := float64(0)
	for _, val := range n {
		total += val
	}
	return total
}
