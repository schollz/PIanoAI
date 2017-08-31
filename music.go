package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Note carries the pitch, velocity, and duration information
// of a single press
type Note struct {
	On       bool
	Pitch    int64
	Velocity int64
	Beat     float64
}

// Time returns when it will be played (or turned off)
func (n *Note) Time() string {
	return fmt.Sprintf("%2.5f", n.Beat)
}

func (n *Note) Name() string {
	return fmt.Sprintf("%d", n.Pitch)
}

// Music stores all the notes that will be played / were already played
type Music struct {
	Notes map[string]map[string]Note // Notes[TIMEPLAYED(on/off)][PITCH] = Note
	sync.RWMutex
}

// NewMusic returns a new object
func NewMusic() *Music {
	m := new(Music)
	m.Lock()
	m.Notes = make(map[string]map[string]Note)
	m.Unlock()
	return m
}

// OpenMusic opens a previous music
func OpenMusic(filename string) (*Music, error) {
	bMusic, err := ioutil.ReadFile(filename)
	if err != nil {
		return new(Music), err
	}
	m := new(Music)
	m.Lock()
	err = json.Unmarshal(bMusic, &m.Notes)
	m.Unlock()
	return m, err
}

// AddNote will add a note in a thread-safe way.
func (m *Music) AddNote(n Note) (err error) {
	m.Lock()
	defer m.Unlock()
	if _, hasTime := m.Notes[n.Time()]; hasTime {
		if _, hasNote := m.Notes[n.Time()][n.Name()]; hasNote {
			return
		}
	} else {
		m.Notes[n.Time()] = make(map[string]Note)
	}
	m.Notes[n.Time()][n.Name()] = n
	return
}

// GetNotes retrieve notes in music in a thread-safe way
func (m *Music) GetNotes(beat float64) (hasNotes bool, notes []Note) {
	m.RLock()
	defer m.RUnlock()
	var notesMap map[string]Note
	notesMap, hasNotes = m.Notes[fmt.Sprintf("%2.5f", beat)]
	if !hasNotes {
		return
	}
	notes = make([]Note, len(notesMap))
	i := 0
	for _, note := range notesMap {
		notes[i] = note
		i++
	}
	return
}

// GetNotes retrieve notes in music in a thread-safe way
func (m *Music) GetAllNotes() (notes []Note) {
	logger := log.WithFields(log.Fields{
		"function": "Music.GetAllNotes",
	})
	m.RLock()
	defer m.RUnlock()
	notes = []Note{}
	for beat := range m.Notes {
		logger.Debug(beat)
		for pitch := range m.Notes[beat] {
			logger.Debug(pitch)
			notes = append(notes, m.Notes[beat][pitch])
		}
	}
	return
}

func (m *Music) Save(filename string) (err error) {
	m.RLock()
	defer m.RUnlock()
	bMusic, err := json.Marshal(m.Notes)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, bMusic, 0755)
}
