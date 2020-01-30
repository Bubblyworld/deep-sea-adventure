package eval

import (
	"testing"

	"github.com/bubblyworld/deep-sea-adventure/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMonteCarlo is just a smoke-screen that ensures that a random player
// will receive at least some utility in a typical game. This is not strictly
// guaranteed to happen, of course, but the probability of failure is very low.
func TestMonteCarlo(t *testing.T) {
	s := state.NewStandardState(6)

	var max float64
	for i := 0; i < 100; i++ {
		util, prob, err := montecarlo(s, 1)
		require.NoError(t, err)
		assert.True(t, prob.Sign() > 0)

		if util > max {
			max = util
		}
	}

	assert.True(t, max > 0)
}

// TestEstimate is just a smoke-screen that ensures that a player will always
// estimate a positive utility from a starting position.
func TestEstimate(t *testing.T) {
	s := state.NewStandardState(6)

	for i := 0; i < 100; i++ {
		util, err := Estimate(s, 1, 100, 10)
		require.NoError(t, err)
		assert.True(t, util > 0)
	}
}
