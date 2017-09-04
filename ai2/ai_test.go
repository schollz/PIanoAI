package ai2

import (
	"fmt"
	"testing"

	"github.com/schollz/pianoai/music"
)

func TestAI2(t *testing.T) {
	ai := New()
	m, err := music.Open("../testing/em_jam.json")
	if err != nil {
		t.Error(err)
	}
	ai.Learn(m)
	fmt.Println("CHORD ARRAY")
	fmt.Println(ai.chordStringArray)
	fmt.Println(ai.Lick(0))
}
