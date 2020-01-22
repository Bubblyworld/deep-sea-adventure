package main

import (
	"math/rand"
	"time"

	"github.com/bubblyworld/deep-sea-adventure/game"
	"github.com/bubblyworld/deep-sea-adventure/state"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Start a game and run it till completion.
	g := game.New([]game.Strategy{
		new(alwaysDeeper), new(alwaysDeeper), new(alwaysDeeper)})

	for {
		g.Run()
		time.Sleep(time.Second)
	}
}

type alwaysDeeper struct{}

func (ad *alwaysDeeper) Turn(s state.State) bool {
	return false // never surrender! always deeper!
}

func (ad *alwaysDeeper) PickUp(s state.State) bool {
	return true // always pick everything up
}

func (ad *alwaysDeeper) Drop(s state.State) (int, bool) {
	return 0, false
}
