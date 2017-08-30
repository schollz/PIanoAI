package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/apex/log"
	"github.com/schollz/jsonstore"
)

// Player is the main structure
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

	// AI STUFF

	// AI stores the AI being used
	AI *MarkovAI
	// BeatsOfSilence waits this number of beats before asking
	// the AI for an improvisation
	BeatsOfSilence        int
	CurrentBeatsOfSilence int
}

// Init initializes the parameters and connects up the piano
func (p *Player) Init(bpm float64, beats ...int) (err error) {
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
	// TODO: Optional to load chord history
	p.ChordsPlaying = new(jsonstore.JSONStore)

	p.AI = new(MarkovAI)

	if len(beats) == 1 {
		p.BeatsOfSilence = beats[0]
	} else {
		p.BeatsOfSilence = 10
	}
	p.CurrentBeatsOfSilence = 0
	return
}

// Close will do the shutdown routines before exiting
func (p *Player) Close() (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Player.Close",
	})
	logger.Info("Closing piano...")
	err = p.Piano.Close()
	if err != nil {
		logger.Error(err.Error())
	}
	return
}

// Start initializes the metronome which emits
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
			logger.Infof("%+v", sig)
			// sig is a ^C, handle it
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
				logger.Infof("beat %2.0f", p.Beat)
			}
			p.Beat += 0.015625
		case <-doneChan:
			fmt.Println("Done")
			return
		}
	}
}
