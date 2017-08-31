package main

import "fmt"
import "sync"

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

// Music stores all the notes that will be played
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

// AddNote will add a note in a thread-safe way.
func (m *Music) AddNote(n Note) {
	m.Lock()
	defer m.Unlock()
	if _, hasTime := m.Notes[n.Time()]; hasTime {
		if _, hasNote := m.Notes[n.Time()][n.Name()]; !hasNote {
			m.Notes[n.Time()][n.Name()] = n
		}
	} else {
		m.Notes[n.Time()] = make(map[string]Note)
	}
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
