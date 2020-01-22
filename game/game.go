// Package game contains utilities for driving a game of deep sea adventure
// where each player has a configured programmable strategy.
package game

import (
	"fmt"
	"math/rand"

	"github.com/bubblyworld/deep-sea-adventure/state"
)

// Strategy encodes the decisions a player needs to make on their turn as a
// programmable interface.
type Strategy interface {
	// Turn is called at the start of the turn if the player hasn't
	// turned around already. If it returns true, the player will begin to
	// retreat back to the submarine.
	Turn(state.State) bool

	// PickUp is called if the player ends on a tile with a treasure. If it
	// returns true, the player will pick the treasure up.
	PickUp(state.State) bool

	// Drop is called if the player ends on an empty tile. If it returns true,
	// the treasure with the given index will be dropped on the tile. If the
	// index is invalid, nothing will happen.
	Drop(state.State) (int, bool)
}

// Game drives a strategy-driven game of deep sea adventure.
type Game struct {
	State      state.State
	Strategies []Strategy // strategy indexed by player
}

func New(sl []Strategy) *Game {
	return &Game{
		State:      state.NewStandardState(len(sl)),
		Strategies: sl,
	}
}

// Run plays through the next player's turn using their configured strategy.
func (g *Game) Run() {
	if g.State.Stage() == state.StageEndOfGame {
		fmt.Printf("game is over!\n")
		return
	}

	s := g.Strategies[g.State.CurrentPlayer()]
	fmt.Printf("BEGIN DFSA STAGE %s\n", g.State.Stage())
	fmt.Printf("%s\n", printState(g.State))
	fmt.Printf("\tround %d, player %d to move (air before turn: %d)\n",
		g.State.Round(), g.State.CurrentPlayer(), g.State.Air())

	var err error
	switch g.State.Stage() {
	case state.StageRoll:
		roll := roll()
		fmt.Printf("\tplayer %d has rolled %d\n",
			g.State.CurrentPlayer(), roll)

		err = g.State.Do(state.Roll(roll))

	case state.StagePickUp:
		pu := s.PickUp(g.State)
		if pu {
			fmt.Printf("\tplayer %d decides to pick up treasure\n",
				g.State.CurrentPlayer())
		} else {
			fmt.Printf("\tplayer %d decides to ignore the treasure\n",
				g.State.CurrentPlayer())
		}

		err = g.State.Do(state.PickUp(pu))

	case state.StageDrop:
		i, d := s.Drop(g.State)
		if d {
			fmt.Printf("\tplayer %d decides to drop treasure %d\n",
				g.State.CurrentPlayer(), i)
		} else {
			fmt.Printf("\tplayer %d decides not to drop anything\n",
				g.State.CurrentPlayer())
		}

		err = g.State.Do(state.Drop(i, d))

	case state.StageTurn:
		t := s.Turn(g.State)
		if t {
			fmt.Printf("\tplayer %d decides to turn around\n",
				g.State.CurrentPlayer())
		} else {
			fmt.Printf("\tplayer %d decides to go deeper\n",
				g.State.CurrentPlayer())
		}

		err = g.State.Do(state.Turn(t))

	default:
		panic("invalid stage reached in game run") // should never happen
	}

	if err != nil {
		panic(fmt.Errorf("error doing decision on game state: %v", err))
	}

	fmt.Printf("END\n\n")
}

func printState(s state.State) string {
	pm := make(map[int]string)
	for i, p := range s.Players() {
		if _, ok := pm[p.Position]; ok {
			pm[p.Position] = "*"
		} else {
			pm[p.Position] = fmt.Sprint(i)
		}
	}

	str := "------------" + fmt.Sprintf("ROUND %d", s.Round()) + "------------\n"
	for i, p := range s.Players() {
		str += fmt.Sprintf("\t player %d: held(%d), stashed(%d)\n", i,
			len(p.HeldTreasure), len(p.StashedTreasure))
	}
	str += "\t"
	for i, t := range s.Tiles() {
		if s, ok := pm[i]; ok {
			str += s
			continue
		}

		switch t.Type {
		case state.TileTypeEmpty:
			str += "."

		case state.TileTypeTreasure:
			str += "$"

		case state.TileTypeSubmarine:
			str += "@"
		}
	}
	str += "\n-------------------------------"

	return str
}

func roll() int {
	return 2 + rand.Intn(3) + rand.Intn(3)
}
