package main

import (
	"math/rand"
	"time"

	"github.com/bubblyworld/deep-sea-adventure/mechanics"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Start a game and run it till completion.
	g := mechanics.NewGame([]mechanics.Strategy{
		new(alwaysDeeper), new(alwaysDeeper), new(alwaysDeeper)})

	for {
		g.Run()
		time.Sleep(time.Second)
	}
}

type alwaysDeeper struct{}

func (ad *alwaysDeeper) TurnAround(p *mechanics.Player,
	gs *mechanics.State) bool {

	return false // never surrender!
}

func (ad *alwaysDeeper) Pickup(p *mechanics.Player,
	gs *mechanics.State) bool {

	return true // always pickup everything
}

func (ad *alwaysDeeper) Drop(p *mechanics.Player,
	gs *mechanics.State) (int, bool) {

	return 0, false
}
