package ai2

import (
	"fmt"
	"testing"

	"github.com/schollz/pianoai/music"
)

func TestAI2(t *testing.T) {
	ai := New(250)
	m, err := music.Open("../testing/em_jam.json")
	if err != nil {
		t.Error(err)
	}
	ai.Learn(m)
	fmt.Println("CHORD ARRAY")
	fmt.Println(ai.chordStringArray)
	fmt.Println(ai.Lick(0))
}

func TestAI1(t *testing.T) {
	ai := New(250)
	m, err := music.Open("../testing/c_scale2.json")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(m.Notes[960])
	ai.Learn(m)
	fmt.Println("CHORD ARRAY")
	fmt.Println(ai.chordStringArray)
	for noteI := range ai.chordStringArray {
		fmt.Println(ai.chordArray[noteI].Pitches)
	}
	// fmt.Println(ai.Lick(0))
}
