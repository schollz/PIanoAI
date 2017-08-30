package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/schollz/jsonstore"
	log "github.com/sirupsen/logrus"
)

// Player is the main structure which facilitates the Piano, and the AI.
// The Player spawns threads for listening to events on the Piano, and also
// spawns threads for playing notes on the piano. It also spawns threads
// for doing the machine learning and using the results.
type Player struct {
	// BPM is the beats per minute
	BPM float64
	// Beat counts the beat number
	Beat float64
	// Key stores the key of the song (TODO: Add in key-signature constraints)
	Key string

	// Piano is the piano that does the playing, the MIDI keyboard
	Piano *Piano
	// ChordsToPlay is a map of future chords to play
	ChordsToPlay *jsonstore.JSONStore
	// ChordHistory is a map of all the previous notes played
	ChordHistory *jsonstore.JSONStore
	// ChordsPlaying is a map of all the chords currently being
	// played
	ChordsPlaying *jsonstore.JSONStore

	// AI stores the AI being used
	AI *MarkovAI
	// BeatsOfSilence waits this number of beats before asking
	// the AI for an improvisation
	BeatsOfSilence float64
	// LastNote is the beat of the last note played
	LastNote float64
}

// Init initializes the parameters and connects up the piano
func (p *Player) Init(bpm float64, beats ...float64) (err error) {

	logger := log.WithFields(log.Fields{
		"function": "Player.Init",
	})
	p.BPM = bpm
	logger.Infof("Initiating player at %2.0f BPM", p.BPM)
	p.Beat = 0
	p.Key = "C"

	p.Piano = new(Piano)
	err = p.Piano.Init()
	if err != nil {
		return
	}

	p.ChordsToPlay = new(jsonstore.JSONStore)
	p.ChordHistory = new(jsonstore.JSONStore)
	// TODO: Option to load chord history
	p.ChordsPlaying = new(jsonstore.JSONStore)

	p.AI = new(MarkovAI)

	if len(beats) == 1 {
		p.BeatsOfSilence = beats[0]
	} else {
		p.BeatsOfSilence = 10
	}
	p.LastNote = 0
	return
}

// Close will do the shutdown routines before exiting
func (p *Player) Close() (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Player.Close",
	})
	logger.Debug("Closing piano...")
	err = p.Piano.Close()
	if err != nil {
		logger.Error(err.Error())
	}
	return
}

// Start initializes the metronome which keeps track of beats
// Each beat will start new threads to Emit new chords, and/or
// generate new Improvisation
func (p *Player) Start() {
	logger := log.WithFields(log.Fields{
		"function": "Player.Start",
	})

	// Exit on Ctl+C
	doneChan := make(chan bool)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			logger.Debugf("%+v", sig)
			// sig is a ^C, handle it
			p.Close()
			doneChan <- true
		}
	}()

	p.Beat = 0
	tickChan := time.NewTicker(time.Millisecond * time.Duration((1000*60/p.BPM)/64)).C
	logger.Infof("Tick size: ~%s", time.Duration(time.Millisecond*time.Duration((1000*60/p.BPM)/64)))
	for {
		select {
		case <-tickChan:
			if p.Beat == math.Trunc(p.Beat) {
				logger.Debugf("beat %2.0f", p.Beat)
			}
			p.Beat += 0.015625
			p.LastNote += 0.015625
			go p.Emit(p.Beat)
			// TODO: If the p.Beat - p.LastNote > p.BeatsOfSilence
			// THEN go p.Improvisation() -> generates new markov model and then generates
			// notes to emit

		case <-doneChan:
			fmt.Println("Done")
			return
		}
	}
}

// Emit will play/stop notes depending on the current beat.
// This should be run in a separate thread.
func (p *Player) Emit(beat float64) {
	// TODO: Check if beat is in any jsonstore for needint go be played
	// If it is, play the note

}

// Listen tells the player to listen to events from the
// piano MIDI connection. This is meant to be run in a
// separate thread.
func (p *Player) Listen() {
	logger := log.WithFields(log.Fields{
		"function": "Player.Listen",
	})

	ch := p.Piano.inputStream.Listen()
	for {
		event := <-ch
		logger.WithFields(log.Fields{
			"timestamp": event.Timestamp,
			"data1":     event.Data1,
			"data2":     event.Data2,
			"status":    event.Status,
		}).Info("key event")
		// TODO: go p.AddEvent()
		// TODO: if the key is ON, then p.LastNote = p.Beat
	}
}
