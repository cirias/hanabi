package hanabi

import (
	"errors"
	"fmt"
)

type Player struct {
	id     int
	game   *Game
	cards  []Card
	events chan interface{}
}

var (
	ErrNotYourTurn      = errors.New("not your turn")
	ErrGameOver         = errors.New("game over")
	ErrInvalidCardIndex = errors.New("card index out of range")
)

func (p *Player) Events() <-chan interface{} {
	return p.events
}

func (p *Player) Aware(t *Player) ([]Card, error) {
	p.game.rwm.RLock()
	defer p.game.rwm.RUnlock()

	if p == t {
		return nil, fmt.Errorf("could not aware yourself")
	}

	cards := make([]Card, len(t.cards))
	copy(cards, t.cards)

	return cards, nil
}

func (p *Player) Cue(t *Player, type_, value int) error {
	p.game.rwm.Lock()
	defer p.game.rwm.Unlock()

	if p.game.over() {
		return ErrGameOver
	}

	if p.id != p.game.currentPlayer {
		return ErrNotYourTurn
	}

	if p.game.infoToken <= 0 {
		return fmt.Errorf("not enough info token")
	}

	cards := make([]int, 0, 5)
	for i, c := range t.cards {
		if c[type_] == value {
			cards = append(cards, i)
		}
	}

	if len(cards) == 0 {
		return fmt.Errorf("no card will be cued")
	}

	p.game.countEndingTurns()
	p.game.nextPlayer()

	p.game.infoToken -= 1

	for _, player := range p.game.Players {
		player.events <- EventCue{
			From:  p.id,
			To:    t.id,
			Cards: cards,
			Type:  type_,
			Value: value,
		}
	}

	return nil
}

func (p *Player) has(i int) bool {
	return i >= 0 && i < len(p.cards)
}

func (p *Player) Play(out, in int) error {
	p.game.rwm.Lock()
	defer p.game.rwm.Unlock()

	if p.game.over() {
		return ErrGameOver
	}

	if p.id != p.game.currentPlayer {
		return ErrNotYourTurn
	}

	if !p.has(out) {
		return ErrInvalidCardIndex
	}

	if !p.has(in) {
		return ErrInvalidCardIndex
	}

	p.game.countEndingTurns()
	p.game.nextPlayer()

	c := p.cards[out]
	p.game.play(c)
	cs := p.game.draw()

	p.cards = append(p.cards[:out], p.cards[out+1:]...)
	p.cards = append(p.cards[:in], append(cs, p.cards[in:]...)...)

	for _, player := range p.game.Players {
		player.events <- EventPlay{
			Player: p.id,
			Card:   c,
		}
	}

	return nil
}

func (p *Player) Discard(out, in int) error {
	p.game.rwm.Lock()
	defer p.game.rwm.Unlock()

	if p.game.over() {
		return ErrGameOver
	}

	if p.id != p.game.currentPlayer {
		return ErrNotYourTurn
	}

	if !p.has(out) {
		return ErrInvalidCardIndex
	}

	if !p.has(in) {
		return ErrInvalidCardIndex
	}

	p.game.countEndingTurns()
	p.game.nextPlayer()

	c := p.cards[out]
	cs := p.game.draw()

	p.cards = append(p.cards[0:out], p.cards[out+1:]...)
	p.cards = append(p.cards[0:in], append(cs, p.cards[in:]...)...)

	if p.game.infoToken < maxInfoToken {
		p.game.infoToken += 1
	}

	for _, player := range p.game.Players {
		player.events <- EventDiscard{
			Player: p.id,
			Card:   c,
		}
	}

	return nil
}

type EventCue struct {
	From  int
	To    int
	Cards []int
	Type  int
	Value int
}

type EventPlay struct {
	Player int
	Card   Card
}

type EventDiscard struct {
	Player int
	Card   Card
}
