// Package state implements the core game state and mechanics.
package state

import (
	"errors"
	"fmt"
	"math/rand"
)

type TreasureType int

const (
	TreasureTypeOne   TreasureType = 1
	TreasureTypeTwo   TreasureType = 2
	TreasureTypeThree TreasureType = 3
	TreasureTypeFour  TreasureType = 4

	// Should always be last, intended as an EOF marker for iteration.
	treasureTypeSentinel TreasureType = 5
)

var treasureValues = map[TreasureType][]int{
	TreasureTypeOne:   []int{0, 0, 1, 1, 2, 2, 3, 3},
	TreasureTypeTwo:   []int{4, 4, 5, 5, 6, 6, 7, 7},
	TreasureTypeThree: []int{8, 8, 9, 9, 10, 10, 11, 11},
	TreasureTypeFour:  []int{12, 12, 13, 13, 14, 14, 15, 15},
}

type Treasure struct {
	Type  TreasureType
	Value int
}

type TreasureStack []Treasure

type TileType int

const (
	TileTypeSubmarine = 1
	TileTypeTreasure  = 2
	TileTypeEmpty     = 3

	// Should always be last, intended as an EOF marker for iteration.
	tileTypeSentinel TileType = 4
)

type Tile struct {
	Type     TileType
	Treasure *TreasureStack // non-nil iff type is TileTypeTreasure
}

type Player struct {
	Position     int
	TurnedAround bool

	// HeldTreasure is the stacks of treasure currently held by the player.
	// Each stack held by the player slows them down by 1 point. If the player
	// survives the end of round (by getting back to the submarine), any held
	// stacks are moved in the stashed stacks list.
	HeldTreasure []TreasureStack

	// StashedTreasure is a list of banked treasure owned by the player. At
	// the end of 3 rounds, the player with the highest total value of stashed
	// treasure wins the game.
	StashedTreasure []TreasureStack
}

type State struct {
	Air     int
	Round   int
	Players []Player
	Tiles   []Tile
}

func New(players []Player) *State {
	res := State{
		Air:     25,
		Round:   1,
		Players: players,
	}

	// The initial map is always laid out as follows. The first tile is the
	// starting submarine. The next 8 tiles are the treasures marked with one
	// dot in random order. Then come the two, three and four dot treasures in
	// similar fashion.
	res.Tiles = append(res.Tiles, Tile{
		Type: TileTypeSubmarine,
	})

	for _, tt := range getTreasureTypes() {
		vl := getTreasureValues(tt)
		rand.Shuffle(len(vl), func(i, j int) {
			vl[i], vl[j] = vl[j], vl[i]
		})

		for _, v := range vl {
			res.Tiles = append(res.Tiles, Tile{
				Type: TileTypeTreasure,
				Treasure: &TreasureStack{
					Treasure{
						Type:  tt,
						Value: v,
					},
				},
			})
		}
	}

	return &res
}

// Validate returns an error if the game state is illegal.
func (gs *State) Validate() error {
	pm := make(map[int]bool)
	for _, p := range gs.Players {
		if !gs.inBounds(p.Position) {
			return errors.New("player is in illegal position")
		}

		if p.Position == 0 {
			continue // multiple players can occupy submarine
		}

		if pm[p.Position] {
			return errors.New("multiple players occupying non-submarine tile")
		}

		pm[p.Position] = true
	}

	return nil
}

// Pickup causes the player to pick-up the treasure stack at their current
// position. If there isn't a treasure to pick up this will error.
func (gs *State) Pickup(p *Player) error {
	if err := gs.Validate(); err != nil {
		return fmt.Errorf("validation error while picking up: %v", err)
	}

	if gs.Tiles[p.Position].Type != TileTypeTreasure {
		return errors.New("player tried to pick up non-treasure tile")
	}

	p.HeldTreasure = append(p.HeldTreasure, *gs.Tiles[p.Position].Treasure)
	gs.Tiles[p.Position] = Tile{
		Type: TileTypeEmpty,
	}

	return nil
}

// Drop causes the player to drop the treasure stack with the given index at
// their current position. If they aren't on an empty tile or the treasure
// stack doesn't exist, this will error.
func (gs *State) Drop(p *Player, index int) error {
	if err := gs.Validate(); err != nil {
		return fmt.Errorf("validation error while dropping: %v", err)
	}

	if index < 0 || index >= len(p.HeldTreasure) {
		return errors.New("player tried to drop non-existent treasure")
	}

	if gs.Tiles[p.Position].Type != TileTypeEmpty {
		return errors.New("player tried to drop on non-empty tile")
	}

	gs.Tiles[p.Position] = Tile{
		Type:     TileTypeTreasure,
		Treasure: &p.HeldTreasure[index],
	}
	p.HeldTreasure = append(p.HeldTreasure[:index],
		p.HeldTreasure[:index+1]...)

	return nil
}

// Move moves the player the given number of spaces, hopping over other
// players as they go. If the player reaches the end of the map, or the
// submarine if they're going backwards, then they stop (no bounceback).
func (gs *State) Move(p *Player, spaces int) error {
	if err := gs.Validate(); err != nil {
		return fmt.Errorf("validation error while moving: %v", err)
	}

	// A map of which squares contain players. Note that we don't have to
	// worry about skipping player 'p' since there's no bounceback.
	sm := make(map[int]bool)
	for _, pl := range gs.Players {
		if pl.Position == 0 {
			continue // multiple players can occupy submarine
		}

		sm[pl.Position] = true
	}

	// Player must always be going forwards unless they've turned around.
	skip := int(1)
	if p.TurnedAround {
		skip = -1
	}
	if spaces < 0 {
		spaces = -spaces
	}

	for i := 0; i < spaces; i++ {
		newPos := p.Position + skip
		for gs.inBounds(newPos) && sm[newPos] {
			newPos += skip // jump over players
		}

		if sm[newPos] {
			return nil // can't actually move, too many players in front
		}

		p.Position = newPos
	}

	return nil
}

func (gs *State) inBounds(pos int) bool {
	return pos >= 0 && pos < len(gs.Tiles)
}

func getTreasureTypes() []TreasureType {
	var res []TreasureType
	for tt := TreasureTypeOne; tt != treasureTypeSentinel; tt++ {
		res = append(res, tt)
	}

	return res
}

func getTreasureValues(tt TreasureType) []int {
	var res []int
	for _, v := range treasureValues[tt] {
		res = append(res, v)
	}

	return res
}
