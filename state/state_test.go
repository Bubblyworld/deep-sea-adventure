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
