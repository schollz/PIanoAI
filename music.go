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
	Pitch    int
	Velocity int
	Beat     int
}

// Time returns when it will be played (or turned off)
func (n *Note) Time() string {
	return fmt.Sprintf("%2.5f", n.Beat)
}

func (n *Note) Name() string {
	return fmt.Sprintf("%d", n.Pitch)
}

// Notes is a structure for sorting the notes based on current beat
type Notes []Note

func (p Notes) Len() int {
	return len(p)
}

func (p Notes) Less(i, j int) bool {
	return p[i].Beat < p[j].Beat
}

func (p Notes) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Music stores all the notes that will be played / were already played
type Music struct {
	// Notes map: tick -> pitch -> note
	Notes map[int]map[int]Note
	sync.RWMutex
}

// NewMusic returns a new object
func NewMusic() *Music {
	m := new(Music)
	m.Lock()
	m.Notes = make(map[int]map[int]Note)
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
	if _, hasTime := m.Notes[n.Beat]; hasTime {
		if _, hasNote := m.Notes[n.Beat][n.Pitch]; hasNote {
			return
		}
	} else {
		m.Notes[n.Beat] = make(map[int]Note)
	}
	m.Notes[n.Beat][n.Pitch] = n
	return
}

// GetNotes retrieve notes in music in a thread-safe way
func (m *Music) GetNotes(beat int) (hasNotes bool, notes []Note) {
	m.RLock()
	defer m.RUnlock()
	var notesMap map[int]Note
	notesMap, hasNotes = m.Notes[beat]
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
