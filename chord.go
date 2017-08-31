package main

import "fmt"

// Note carries the pitch, velocity, and duration information
// of a single press
type Note struct {
	On       bool
	Pitch    int64
	Velocity int64
	Start    float64
	Stop     float64
}

// IDs must be unique in the song
func (n *Note) ID() string {
	if n.On {
		return fmt.Sprintf("%d-%d", n.Pitch, n.Start)
	} else {
		return fmt.Sprintf("%d-%d", n.Pitch, n.End)
	}
}

func (n *Note) String() string {
	return fmt.Sprintf("Pitch: %d Velocity: %d Duration: %d", n.Pitch, n.Velocity, n.Duration)
}

type Music struct {
	Notes map[float64]map[int64]Note // Notes[TIMEPLAYED(on/off)][PITCH] = Note
}
