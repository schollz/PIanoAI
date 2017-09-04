package ai

import (
	"errors"
	"math/rand"

	"github.com/schollz/gobrain"
	"github.com/schollz/pianoai/music"
	log "github.com/sirupsen/logrus"
)

// Learn is for calculating the matricies for the transition
// probabilities
func (ai *AI) Learn3(notes music.Notes) (err error) {
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

	for j := 0; j <= 3; j++ {
		patterns := [][][]float64{}
		for i, note := range ai.notes {
			if i == 0 {
				continue
			}
			previousNote := convertIntsToFloats([]int{ai.notes[i-1][j], int(rand.Int31())})
			currentNote := convertIntsToFloats([]int{note[j], int(rand.Int31())})
			pattern := [][]float64{
				previousNote, currentNote,
			}
			patterns = append(patterns, pattern)
		}
		// instantiate the Feed Forward
		ai.ff2[j] = &gobrain.FeedForward{}
		logger.Debug("Initializing neural net...")
		ai.ff2[j].Init(2*32, 10, 2*32)
		logger.Debug("Training neural net...")
		ai.ff2[j].Train(patterns, 5000, 0.1, 0.4, true)
		logger.Debug("Finished training.")
	}
	ai.HasLearned = true
	return
}

// Lick2 generates a sequence of chords using the Markov
// probabilities. Must run Learn2() beforehand.
func (ai *AI) Lick3(startBeat int) (lick *music.Music, err error) {
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
		note = []int{0, 0, 0, 0}
		for j := 0; j <= 3; j++ {
			noteInput := convertIntsToFloats([]int{note[j], int(rand.Int31())})
			noteOutput := ai.ff2[j].Update(noteInput)
			note[j] = convertFloatToInts(noteOutput)[0]
		}
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
