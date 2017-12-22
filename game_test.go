package hanabi

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestGame(t *testing.T) {
	rand.Seed(0)
	g, err := NewGame(5)
	if err != nil {
		t.Fatalf("could not new game: %s", err)
	}

	go func() {
		for {
			for _, p := range g.Players {
				e := <-p.Events()
				fmt.Println(p.id, e)
				// fmt.Println(p.cards)
			}
		}
	}()

	for _, p := range g.Players {
		fmt.Println(p.id, p.cards)
	}

	fmt.Println(g.Capture())
	if err := g.Players[0].Play(3, 0); err != nil {
		t.Fatalf("could not play: %s", err)
	}
	fmt.Println(g.Capture())
	fmt.Println(g.Players[0].cards)

	err = g.Players[0].Cue(g.Players[1], CardColor, White)
	if err != ErrNotYourTurn {
		t.Error("should return error:", ErrNotYourTurn)
	}

	fmt.Println(g.Players[1].cards)
	err = g.Players[1].Discard(2, 3)
	if err != nil {
		t.Fatalf("could not discard: %s", err)
	}
	fmt.Println(g.Players[1].cards)
}
