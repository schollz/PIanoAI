package ai

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/schollz/rpiai-piano/music"
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

	note := ai.notes[rand.Intn(len(ai.notes))]
	for j := 0; j < 10; j++ {
		note = ai.GenerateNote(note)
		fmt.Println(note)
	}
	fmt.Println(ai.matrices[1])

	fmt.Println("---LICK---")
	fmt.Println(ai.Lick(0))

	// fmt.Println(ai.matrices[0])
	// fmt.Println(ai.matrices[0][-200][-200])

	// ai.Learn2(m.GetAll())
	// fmt.Println("---LICK---")
	// fmt.Println(ai.Lick2(0))

}
