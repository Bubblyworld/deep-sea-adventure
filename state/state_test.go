package state_test

import (
	"testing"

	"github.com/bubblyworld/deep-sea-adventure/state"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name string
		pl   []int
		err  bool
	}{
		{
			name: "OK - everyone on starting square",
			pl:   []int{0, 0, 0},
		},
		{
			name: "OK - everyone on different squares",
			pl:   []int{0, 1, 2, 3, 4, 5},
		},
		{
			name: "BAD - invalid squares under",
			pl:   []int{0, 1, 2, 3, 4, -5},
			err:  true,
		},
		{
			name: "BAD - invalid squares over",
			pl:   []int{0, 1, 2, 3, 4, 1005},
			err:  true,
		},
		{
			name: "BAD - multiple people on same square",
			pl:   []int{0, 1, 2, 3, 4, 4},
			err:  true,
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			gs := newState(c.pl)
			if c.err {
				assert.Error(t, gs.Validate())
			} else {
				assert.NoError(t, gs.Validate())
			}
		})
	}
}

func TestMove(t *testing.T) {
	cases := []struct {
		name  string
		pl    []int
		turns []bool
		moves []int
		expPl []int
	}{
		{
			name:  "first moves",
			pl:    []int{0, 0},
			turns: []bool{false, false},
			moves: []int{1, 1},
			expPl: []int{1, 2},
		},
		{
			name:  "midgame moves",
			pl:    []int{4, 2, 3},
			turns: []bool{false, false, false},
			moves: []int{0, 1, 4},
			expPl: []int{4, 5, 9},
		},
		{
			name:  "stops at the end",
			pl:    []int{0},
			turns: []bool{false},
			moves: []int{1000},
			expPl: []int{32}, // 4x8 treasure tiles at start
		},
		{
			name:  "stops at start",
			pl:    []int{2},
			turns: []bool{true},
			moves: []int{1000},
			expPl: []int{0},
		},
		{
			name:  "one player in reverse",
			pl:    []int{1, 2, 5},
			turns: []bool{false, false, true},
			moves: []int{2, 1, 3},
			expPl: []int{4, 3, 0},
		},
		{
			name:  "stuck at end if there's a lot of traffic",
			pl:    []int{32, 31, 30, 29},
			turns: []bool{false, false, false, false},
			moves: []int{10, 10, 10, 10},
			expPl: []int{32, 31, 30, 29},
		},
		{
			name:  "not stuck at beginning if there's a lot of traffic",
			pl:    []int{5, 4, 3, 2, 1, 0},
			turns: []bool{true, true, true, true, true, true},
			moves: []int{10, 10, 10, 10, 10, 10},
			expPl: []int{0, 0, 0, 0, 0, 0},
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			gs := newState(c.pl)
			for i, t := range c.turns {
				gs.Players[i].TurnedAround = t
			}
			for i, s := range c.moves {
				assert.NoError(t, gs.Move(&gs.Players[i], s))
			}

			assertPositions(t, gs, c.expPl)
		})
	}
}

func newState(pl []int) *state.State {
	var pll []state.Player
	for _, p := range pl {
		pll = append(pll, state.Player{
			Position: p,
		})
	}

	return state.New(pll)
}

func assertPositions(t *testing.T, gs *state.State, pl []int) {
	for i, p := range gs.Players {
		assert.Equal(t, pl[i], p.Position)
	}
}
