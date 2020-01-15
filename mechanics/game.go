package mechanics

import (
	"log"
	"math/rand"
)

type Strategy interface {
	// TurnAround is called at the start of the turn if the player hasn't
	// turned around already. If it returns true, the player will begin to
	// retreat back to the submarine.
	TurnAround(*Player, *State) bool

	// Pickup is called if the player ends on a tile with a treasure. If it
	// returns true, the player will pick the treasure up.
	Pickup(*Player, *State) bool

	// Drop is called if the player ends on an empty tile. If it returns true,
	// the treasure with the given index will be dropped on the tile. If the
	// index is invalid, nothing will happen.
	Drop(*Player, *State) (int, bool)
}

type Game struct {
	Index      int // which player is next to move
	State      *State
	Strategies []Strategy // strategy indexed by player
}

func NewGame(sl []Strategy) *Game {
	var pl []Player
	for range sl {
		pl = append(pl, Player{})
	}

	return &Game{
		Index:      0,
		State:      New(pl),
		Strategies: sl,
	}
}

// Run plays through the next player's turn using their configured strategy.
func (g *Game) Run() {
	if g.State.Turn >= 4 {
		log.Printf("game is over!")
		return
	}

	if g.State.Air <= 0 {
		log.Printf("round is over, ending and resetting for next round")
		if err := g.State.EndTurn(); err != nil {
			panic(err)
		}

		g.Index = 0 // TODO: furthest player from sub should go first
		return
	}

	p := &g.State.Players[g.Index]
	s := g.Strategies[g.Index]
	if p.TurnedAround && p.Position == 0 {
		log.Printf("round %d, skipping player %d as they have survived",
			g.State.Turn, g.Index)

		g.Index = (g.Index + 1) % len(g.State.Players)
		return // player has already finished their round
	}

	prevAir := g.State.Air
	g.State.Air -= len(p.HeldTreasure)
	if g.State.Air < 0 {
		g.State.Air = 0
	}

	log.Printf("round %d, player %d to move (air %d -> %d)",
		g.State.Turn, g.Index, prevAir, g.State.Air)

	if !p.TurnedAround {
		p.TurnedAround = s.TurnAround(p, g.State)

		if p.TurnedAround {
			log.Printf("\tplayer %d decides to turn around", g.Index)
		}
	} else {
		log.Printf("\tplayer %d has already turned around", g.Index)
	}

	roll := roll()
	log.Printf("\tplayer %d has rolled %d", g.Index, roll)
	if err := g.State.Move(p, roll); err != nil {
		panic(err)
	}

	switch g.State.Tiles[p.Position].Type {
	case TileTypeEmpty:
		if index, ok := s.Drop(p, g.State); ok {
			log.Printf("\tplayer %d decides to drop treasure", g.Index)

			if err := g.State.Drop(p, index); err != nil {
				panic(err)
			}
		}

	case TileTypeTreasure:
		if s.Pickup(p, g.State) {
			log.Printf("\tplayer %d decides to pick up treasure", g.Index)

			if err := g.State.Pickup(p); err != nil {
				panic(err)
			}
		}
	}

	g.Index = (g.Index + 1) % len(g.State.Players)
}

func roll() int {
	return 2 + rand.Intn(3) + rand.Intn(3)
}
