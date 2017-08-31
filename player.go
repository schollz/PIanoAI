package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
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
	var errOpen error
	p.ChordHistory, errOpen = jsonstore.Open("chord_history.json")
	if errOpen != nil {
		logger.WithFields(log.Fields{
			"msg": "Could not open chord history, making new",
		}).Warn(errOpen.Error())
		p.ChordHistory = new(jsonstore.JSONStore)
	} else {
		logger.Infof("Loaded %d chords from history", len(p.ChordHistory.Keys()))
	}

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

	// start listening
	go p.Listen()

	p.Beat = 0
	tickChan := time.NewTicker(time.Millisecond * time.Duration((1000*60/p.BPM)/64)).C
	logger.Infof("Tick size: ~%s", time.Duration(time.Millisecond*time.Duration((1000*60/p.BPM)/64)))
	for {
		select {
		case <-tickChan:
			// if p.Beat == math.Trunc(p.Beat) {
			// 	logger.Debugf("beat %2.0f", p.Beat)
			// }
			p.Beat += 0.015625
			p.LastNote += 0.015625
			go p.Emit(p.Beat)

			if p.Beat-p.LastNote > p.BeatsOfSilence {
				go p.Improvisation()
			}

		case <-doneChan:
			fmt.Println("Done")
			return
		}
	}
}

// Improvisation generates an improvisation from the AI
// and loads into the next beats to be playing
func (p *Player) Improvisation() {

}

// Emit will play/stop notes depending on the current beat.
// This should be run in a separate thread.
func (p *Player) Emit(beat float64) {
	beatStr := strconv.FormatFloat(beat, 'E', -1, 64)
	var chordToPlay Chord
	err := p.ChordsToPlay.Get(beatStr, &chordToPlay)
	if err == nil {
		// TODO: USE SYNC FOR PIANO?
		p.Piano.PlayChord(chordToPlay, p.BPM)
		p.LastNote = beat
	}
}

// AddChordToPlay will add chords to be played and
// to be silenced
func (p *Player) AddChordToPlay(c Chord) (err error) {
	start := time.Now()
	logger := log.WithFields(log.Fields{
		"function": "Player.AddChordToPlay",
	})
	var c2 Chord

	// handle first notes
	onChord := Chord{Notes: []Note{}, Start: c.Start}
	for _, note := range c.Notes {
		logger.WithFields(log.Fields{
			"p": note.Pitch,
			"d": note.Duration,
		}).Debug("adding start note")
		if note.Velocity > 0 {
			onChord.Notes = append(onChord.Notes, note)
		}
	}
	startString := strconv.FormatFloat(c.Start, 'E', -1, 64)
	errGet := p.ChordsToPlay.Get(startString, &c2)
	if errGet == nil {
		logger.Debug("Combining notes")
		onChord.Notes = []Note{}
		for _, note := range c.Notes {
			hasNote := false
			for _, note2 := range c2.Notes {
				if note2.Velocity == note.Velocity && note2.Pitch == note.Pitch {
					hasNote = true
					break
				}
			}
			if !hasNote {
				onChord.Notes = append(onChord.Notes, note)
			}
		}

		c2.Notes = append(c2.Notes, onChord.Notes...)
		err = p.ChordsToPlay.Set(startString, c2)
		if err != nil {
			return
		}
	} else {
		err = p.ChordsToPlay.Set(startString, onChord)
		if err != nil {
			return
		}
	}

	// handle finish notes
	for _, note := range c.Notes {
		logger.WithFields(log.Fields{
			"p": note.Pitch,
			"d": note.Duration,
		}).Debug("adding finish note")
		note.Velocity = 0
		offChord := Chord{Notes: []Note{note}, Start: c.Start + note.Duration}
		startString := strconv.FormatFloat(offChord.Start, 'E', -1, 64)
		errGet = p.ChordsToPlay.Get(startString, &c2)
		if errGet == nil {
			hasNote := false
			for _, note2 := range c2.Notes {
				if note2.Velocity == note.Velocity && note2.Pitch == note.Pitch {
					hasNote = true
					break
				}
			}
			if hasNote {
				logger.Debug("Skipping, already have note")
				continue
			}
			logger.Debug("Combining notes")
			c2.Notes = append(c2.Notes, offChord.Notes...)
			err = p.ChordsToPlay.Set(startString, c2)
			if err != nil {
				return
			}
		} else {
			err = p.ChordsToPlay.Set(startString, offChord)
			if err != nil {
				return
			}
		}
	}
	logger.Debug("Took %s", time.Since(start).String())
	return
}

// AddChordToHistory will add or merge the specified chord into
// the history
func (p *Player) AddChordToHistory(c Chord) (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Player.AddChordToHistory",
	})
	var c2 Chord
	historyString := strconv.FormatFloat(c.Start, 'E', -1, 64)
	errGet := p.ChordHistory.Get(historyString, &c2)
	if errGet == nil {
		logger.Debug("Combining notes")
		c.Notes = append(c.Notes, c2.Notes...)
	}
	for _, note := range c.Notes {
		logger.WithFields(log.Fields{
			"pitch":    note.Pitch,
			"duration": note.Duration,
		}).Debug("adding to history")
	}
	err = p.ChordHistory.Set(historyString, c)
	if err != nil {
		logger.Error(err.Error())
	}
	return
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
			"pitch":    event.Data1,
			"on":       event.Data2 == 0,
			"velocity": event.Data2,
		}).Info("key event")
		// TODO: Cast the data to a Note
		note := Note{
			Pitch:    event.Data1,
			Velocity: event.Data2,
			Duration: p.Beat, // save the current beat in the Duration
		}

		if note.Pitch == 21 {
			logger.Info("Saving chord_history.json")
			go jsonstore.Save(p.ChordHistory, "chord_history.json")
			continue
		} else if note.Pitch == 22 {
			// Play entire chord history
			for _, key := range p.ChordHistory.Keys() {
				var chord Chord
				p.ChordHistory.Get(key, &chord)
				p.AddChordToPlay(chord)
				p.Beat = 0
			}
			continue
		}

		playingString := strconv.Itoa(int(note.Pitch))
		if note.Velocity == 0 {
			p.LastNote = p.Beat
			var note2 Note
			errGet := p.ChordsPlaying.Get(playingString, &note2)
			if errGet == nil {
				// determine the actual duration
				note.Duration = note.Duration - note2.Duration
				// use the previous velocity
				note.Velocity = note2.Velocity
				chord := Chord{
					Notes: []Note{note},
					Start: note2.Duration, // the previous note duration still contains the beat, not the duration
				}
				errAdd := p.AddChordToHistory(chord)
				if errAdd != nil {
					logger.Warn(errAdd.Error())
				}
				p.ChordsPlaying.Delete(playingString)

			} else {
				logger.Warn(errGet.Error())
			}
		} else {
			logger.Debugf("Adding '%s' to ChordsPlaying", playingString)
			errAdd := p.ChordsPlaying.Set(playingString, note)
			if errAdd != nil {
				logger.Warn(errAdd.Error())
			}
		}

	}
}
