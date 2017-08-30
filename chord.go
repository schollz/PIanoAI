package main

import "fmt"

// Note carries the pitch, velocity, and duration information
// of a single press
type Note struct {
	Pitch    int64
	Velocity int64
	Duration float64 // Duration is in beats
}

func (n *Note) String() string {
	return fmt.Sprintf("Pitch: %d Velocity: %d Duration: %d", n.Pitch, n.Velocity, n.Duration)
}

// Chord contains information one or multiple notes.
// Chords have a start, in which all the notes are started.
// After the start, the duration of each note depends on
// each individual note.
type Chord struct {
	Notes []Note
	Start float64 // Start is in beats
}

func (c *Chord) String() string {
	s := fmt.Sprintf("%d Notes", len(c.Notes))
	for _, note := range c.Notes {
		s += note.String()
	}
	return s
}
