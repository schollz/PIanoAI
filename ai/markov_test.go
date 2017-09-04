package ai

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/schollz/pianoai/music"
)

func TestAI1(t *testing.T) {
	ai := New()
	m, err := music.Open("../testing/c_scale.json")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(m.GetAll())
	analyzedNotes := ai.Analyze(m.GetAll())
	fmt.Println(analyzedNotes)
	fmt.Println(ai)
	ai.Learn(m.GetAll())
	fmt.Println(ai.matrices[0])
	fmt.Println(analyzedNotes)

	fmt.Println(pickRandom(ai.matrices[0][65][-1]))

	noteIndex := rand.Intn(len(ai.notes)-1) + 1
	note1 := ai.notes[noteIndex]
	note2 := ai.notes[noteIndex-1]
	for j := 0; j < 10; j++ {
		note := ai.GenerateNote(note1, note2)
		fmt.Println(note)
		note2 = note1
		note1 = note
	}
	fmt.Println(ai.matrices[1])

	fmt.Println("---LICK---")
	fmt.Println(ai.Lick(0))

	// fmt.Println(ai.matrices[0])
	// fmt.Println(ai.matrices[0][-200][-200])

	ai.Learn4(m.GetAll())
	fmt.Println("---LICK---")
	fmt.Println(ai.Lick4(0))

}

func TestAI2(t *testing.T) {
	ai := New()
	m, err := music.Open("../testing/em_jam.json")
	if err != nil {
		t.Error(err)
	}
	ai.Learn(m.GetAll())
	fmt.Println(ai.matrices[0])

	noteIndex := rand.Intn(len(ai.notes)-1) + 1
	note1 := ai.notes[noteIndex]
	note2 := ai.notes[noteIndex-1]
	for j := 0; j < 10; j++ {
		note := ai.GenerateNote(note1, note2)
		fmt.Println(note)
		note2 = note1
		note1 = note
	}
	fmt.Println(ai.matrices[0])
	fmt.Println(ai.matrices[0][88])

	fmt.Println("---LICK---")
	fmt.Println(ai.Lick(0))
	fmt.Println(ai.notes)
}
