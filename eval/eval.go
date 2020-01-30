// Package eval contains utilities for evaluating deep sea adventure positions
// based on alpha/beta pruning and monte-carlo utility estimations.
package eval

import (
	"math/big"
	"math/rand"

	"github.com/bubblyworld/deep-sea-adventure/state"
)

// Number of random monte-carlo games to consider for utility estimation.
const estimateIterations = 100

// Evaluate returns a map of valid decisions to their approximate expected
// utility for the current player. To minimise exploitability we assume that
// other players are working in tandem to undermine the given player, and only
// compute till the end of the current round to avoid evaluation drifts over
// longer-term computation.
// TODO: This assumption may be too strict - we should probably assume each
//       player has their interests at heart.
func Evaluate(s state.State, depth int) (
	map[state.Decision]float64, error) {

	return evaluate(s, s.CurrentPlayer(), depth, s.Round()+1)
}

func evaluate(s state.State, player, depth, lastRound int) (
	map[state.Decision]float64, error) {

	if depth <= 0 || s.Round() >= lastRound {
		return nil, nil // we're done, at max depth
	}

	vdl := s.ValidDecisions()
	if len(vdl) == 0 {
		return nil, nil // the game is over
	}

	dm := make(map[state.Decision]float64)
	for _, vd := range vdl {
		if err := s.Do(vd); err != nil {
			return nil, err
		}

		cdm, err := evaluate(s, player, depth-1, lastRound)
		if err != nil {
			return nil, err
		}

		// If we get nil from the child, it means we've either reached the
		// end of the game or we've bottomed out depth-wise. In this case we
		// need to pass to the cheaper monte-carlo estimation to avoid the
		// crazy combinatorial explosion of states.
		var best float64
		if len(cdm) == 0 {
			e, err := Estimate(s, player, estimateIterations, lastRound)
			if err != nil {
				return nil, err
			}

			best = e
		} else {
			cp := s.CurrentPlayer()
			best = float64(-999999) // "negative infinity" X_X
			if cp != player {
				best = float64(999999) // "positive infinity" 0_D
			}

			for _, eval := range cdm {
				// Try to maximise the score if we're the given player, else
				// we try and screw them over as hard as possible.
				if cp == player {
					if eval > best {
						best = eval
					}
				} else {
					if eval < best {
						best = eval
					}
				}
			}
		}

		dm[vd] = best
		if err := s.Undo(); err != nil {
			return nil, err
		}
	}

	return dm, nil
}

// Estimate returns an estimate for the expected utility for the given player
// in the given board state. Calculations are performed using weighted monte-
// -carlo tree searches to possible end states.
func Estimate(s state.State, player, iterations, lastRound int) (
	float64, error) {

	// If the game is already over, we can actually be exact.
	if s.Round() >= lastRound || s.Stage() == state.StageEndOfGame {
		return rawUtility(s, player), nil
	}

	var ul []float64
	var pl []*big.Rat
	psum := new(big.Rat)
	for i := 0; i < iterations; i++ {
		util, prob, err := montecarlo(s, player)
		if err != nil {
			return 0, err
		}

		ul = append(ul, util)
		pl = append(pl, prob)
		psum.Add(psum, prob)
	}

	var usum float64
	for i := 0; i < iterations; i++ {
		prob := new(big.Rat)
		prob.Quo(pl[i], psum)
		probf, _ := prob.Float64()

		usum += probf * ul[i]
	}

	return usum, nil
}

var diceProbability = map[int]*big.Rat{
	2: big.NewRat(1, 9),
	3: big.NewRat(2, 9),
	4: big.NewRat(3, 9),
	5: big.NewRat(2, 9),
	6: big.NewRat(1, 9),
}

func montecarlo(s state.State, player int) (float64, *big.Rat, error) {
	if s.Stage() == state.StageEndOfGame { // end of game
		return rawUtility(s, player), big.NewRat(1, 1), nil
	}

	vdl := s.ValidDecisions()
	vd := vdl[rand.Intn(len(vdl))]
	prob := big.NewRat(1, int64(len(vdl)))
	if s.Stage() == state.StageRoll { // the only chance nodes are rolls
		prob = diceProbability[int(vd.Value())]
	}

	if err := s.Do(vd); err != nil {
		return 0, nil, err
	}

	util, childProb, err := montecarlo(s, player)
	if err != nil {
		return 0, nil, err
	}

	if err := s.Undo(); err != nil {
		return 0, nil, err
	}

	rp := new(big.Rat)
	return util, rp.Mul(prob, childProb), nil
}

var expectedUtility = map[state.TreasureType]float64{
	state.TreasureTypeOne:   1.5,
	state.TreasureTypeTwo:   5.5,
	state.TreasureTypeThree: 9.5,
	state.TreasureTypeFour:  13.5,
}

func rawUtility(s state.State, player int) float64 {
	return sum(s.Players()[player].StashedTreasure)
}

func sum(tsl []state.TreasureStack) float64 {
	var sum float64
	for _, ts := range tsl {
		for _, t := range ts {
			sum += expectedUtility[t.Type]
		}
	}

	return sum
}
