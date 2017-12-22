package hanabi

import (
	"fmt"
	"math/rand"
	"sync"
)

const (
	maxInfoToken = 8
	maxFuseToken = 3
)

const (
	CardColor = iota
	CardNumber
)

const (
	White = iota
	Yellow
	Green
	Blue
	Red
)

type Card [2]int

func (c Card) String() string {
	return fmt.Sprintf("%c%d", 'a'+c[CardColor], c[CardNumber])
}

type Game struct {
	rwm           sync.RWMutex
	deck          []Card
	played        [5]int
	infoToken     int
	fuseToken     int
	currentPlayer int
	endingTurns   int
	Players       []*Player
}

func NewGame(playerCount int) (*Game, error) {
	cardsPerNumber := []int{3, 2, 2, 2, 1}

	deck := make([]Card, 0)
	for color := 0; color < 5; color += 1 {
		for number, n := range cardsPerNumber {
			for i := 0; i < n; i += 1 {
				c := Card{}
				c[CardColor] = color
				c[CardNumber] = number + 1
				deck = append(deck, c)
			}
		}
	}

	// rand.Seed(time.Now().Unix())
	for i := range deck {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}

	if playerCount < 2 || playerCount > 5 {
		return nil, fmt.Errorf("invalid player count %d", playerCount)
	}

	cardsPerPlayer := 4
	if playerCount < 4 {
		cardsPerPlayer = 5
	}

	game := &Game{
		infoToken: maxInfoToken,
		fuseToken: maxFuseToken,
	}

	players := make([]*Player, playerCount)
	for i, _ := range players {
		players[i] = &Player{
			id:     i,
			game:   game,
			cards:  deck[:cardsPerPlayer],
			events: make(chan interface{}),
		}
		deck = deck[cardsPerPlayer:]
	}

	game.deck = deck
	game.Players = players

	return game, nil
}

func (g *Game) Over() bool {
	g.rwm.RLock()
	defer g.rwm.RUnlock()

	return g.over()
}

func (g *Game) over() bool {
	if g.endingTurns >= len(g.Players) {
		return true
	}

	if g.fuseToken <= 0 {
		return true
	}

	allFive := true
	for _, n := range g.played {
		allFive = allFive && n == 5
	}
	if allFive {
		return true
	}

	return false
}

func (g *Game) countEndingTurns() {
	if len(g.deck) == 0 {
		g.endingTurns += 1
	}
}

func (g *Game) nextPlayer() {
	g.currentPlayer += 1
	g.currentPlayer %= len(g.Players)
}

func (g *Game) play(c Card) {
	color := c[CardColor]
	number := c[CardNumber]
	if g.played[color]+1 != number {
		g.fuseToken -= 1
		return
	}

	g.played[color] = number
	if number == 5 && g.infoToken < maxInfoToken {
		g.infoToken += 1
	}
}

func (g *Game) draw() []Card {
	if len(g.deck) == 0 {
		return []Card{}
	}

	cs := g.deck[0:1]
	g.deck = g.deck[1:]

	return cs
}

type Snapshot struct {
	Played        [5]int
	InfoToken     int
	FuseToken     int
	DeckLength    int
	CurrentPlayer int
}

func (g *Game) Capture() Snapshot {
	g.rwm.RLock()
	defer g.rwm.RUnlock()

	var played [5]int
	copy(played[:], g.played[:])

	return Snapshot{
		Played:        played,
		InfoToken:     g.infoToken,
		FuseToken:     g.fuseToken,
		DeckLength:    len(g.deck),
		CurrentPlayer: g.currentPlayer,
	}
}
