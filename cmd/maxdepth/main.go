// Package main is an experiment to determine the max decision depth of the
// first round of deep sea adventure. The assumption is that this will be
// maximised in the first round since there are fewer treasures on the board
// in later rounds.
//
// A "decision" is defined to be:
//  * rolling the dice (even though this is a chance node, technically)
//  * deciding whether to pick up
//  * deciding whether to drop
//  * deciding whether to turn around
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/bubblyworld/deep-sea-adventure/state"
)

var players = flag.Int("players", 6,
	"number of players to compute maximum depth for")

var profile = flag.String("profile", "",
	"file to write CPU profile to if desired")

func main() {
	flag.Parse()

	s := state.NewStandardState(*players)
	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			panic(err)
		}

		pprof.StartCPUProfile(f)
		go func() {
			time.Sleep(time.Second * 30)
			pprof.StopCPUProfile()
		}()
	}

	max, err := do(s, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Max depth achieved: %d\n", max)
}

func do(s state.State, depth int) (int, error) {
	if s.Round() > 1 { // we only care about a single round
		return depth, nil
	}

	dl := s.ValidDecisions()
	max := depth
	for _, d := range dl {
		if err := s.Do(d); err != nil {
			return 0, err
		}

		childMax, err := do(s, depth+1)
		if err != nil {
			return 0, err
		}
		if childMax > max {
			max = childMax
		}

		if err := s.Undo(); err != nil {
			return 0, err
		}
	}

	return max, nil
}
