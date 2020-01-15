package mechanics

import (
	"fmt"
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
		fmt.Printf("game is over!\n")
		return
	}

	if g.State.Air <= 0 {
		fmt.Printf("round is over, ending and resetting for next round\n")
		if err := g.State.EndTurn(); err != nil {
			panic(err)
		}

		g.Index = 0 // TODO: furthest player from sub should go first
		return
	}

	p := &g.State.Players[g.Index]
	s := g.Strategies[g.Index]
	if p.TurnedAround && p.Position == 0 {
		fmt.Printf("round %d, skipping player %d as they have survived\n",
			g.State.Turn, g.Index)

		g.Index = (g.Index + 1) % len(g.State.Players)
		return // player has already finished their round
	}

	prevAir := g.State.Air
	g.State.Air -= len(p.HeldTreasure)
	if g.State.Air < 0 {
		g.State.Air = 0
	}

	fmt.Printf("\n%s\n", g.State)
	fmt.Printf("round %d, player %d to move (air %d -> %d)\n",
		g.State.Turn, g.Index, prevAir, g.State.Air)

	if !p.TurnedAround {
		p.TurnedAround = s.TurnAround(p, g.State)

		if p.TurnedAround {
			fmt.Printf("\tplayer %d decides to turn around\n", g.Index)
		}
	} else {
		fmt.Printf("\tplayer %d has already turned around\n", g.Index)
	}

	roll := roll()
	moves := roll - len(p.HeldTreasure)
	if moves < 0 {
		moves = 0
	}

	fmt.Printf("\tplayer %d has rolled %d, moving %d\n", g.Index, roll, moves)
	if err := g.State.Move(p, moves); err != nil {
		panic(err)
	}

	switch g.State.Tiles[p.Position].Type {
	case TileTypeEmpty:
		if index, ok := s.Drop(p, g.State); ok {
			fmt.Printf("\tplayer %d decides to drop treasure\n", g.Index)

			if err := g.State.Drop(p, index); err != nil {
				panic(err)
			}
		}

	case TileTypeTreasure:
		if s.Pickup(p, g.State) {
			fmt.Printf("\tplayer %d decides to pick up treasure\n", g.Index)

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
