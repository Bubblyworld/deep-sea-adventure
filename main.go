package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

	"github.com/bubblyworld/deep-sea-adventure/game"
	"github.com/bubblyworld/deep-sea-adventure/state"
)

var profile = flag.String("profile", "", "where to save CPU profile")

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		go func() {
			time.Sleep(time.Second * 60)

			pprof.StopCPUProfile()
		}()
	}

	// Start a game and run it till completion.
	g := game.New([]game.Strategy{
		new(alwaysDeeper), new(alwaysDeeper), new(alwaysDeeper)})

	for {
		if g.State.Stage() == state.StageEndOfGame {
			fmt.Println("Thanks for watching! Bye now.")
			break
		}

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
