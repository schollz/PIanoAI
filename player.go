package main

import "github.com/schollz/jsonstore"

type Player struct {
	// BPM is the beats per minute
	BPM int
	// Tick counts up from 0 in intervals of 1/64 beats
	Tick int64
	// Key stores the key of the song (TODO: Add in key-signature constraints)
	Key string

	// Piano is the piano that does the playing, the MIDI keyboard
	Piano *Piano
	// ChordsToPlay is a map of future chords to play
	ChordsToPLay *jsonstore.JSONStore
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
