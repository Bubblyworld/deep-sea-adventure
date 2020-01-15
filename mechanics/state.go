// Package mechanics implements the core game state and mechanics.
package mechanics

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
	TileTypeSubmarine TileType = 1
	TileTypeTreasure  TileType = 2
	TileTypeEmpty     TileType = 3

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
	Turn    int
	Players []Player
	Tiles   []Tile
}

// TODO: Change the input to an int. Kinda silly right now.
func New(players []Player) *State {
	res := State{
		Air:     25,
		Turn:    1,
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
		if !gs.inBounds(newPos) {
			return nil // already at the end
		}
		for gs.inBounds(newPos) && sm[newPos] {
			newPos += skip // jump over players
		}

		if !gs.inBounds(newPos) || sm[newPos] {
			return nil // can't actually move, too many players in front
		}

		p.Position = newPos
	}

	return nil
}

// EndTurn kills any players that have yet to reach the submarine and resets
// the state for the next turn.
func (gs *State) EndTurn() error {
	if err := gs.Validate(); err != nil {
		return fmt.Errorf("validation error while ending turn: %v", err)
	}

	// Reset all the players, keeping treasure if they survived.
	var tl []Treasure
	for i := range gs.Players {
		p := &gs.Players[i]
		survived := p.Position == 0
		p.Position = 0
		p.TurnedAround = false

		if survived {
			p.StashedTreasure = append(p.StashedTreasure, p.HeldTreasure...)
			p.HeldTreasure = nil
		} else {
			for _, t := range p.HeldTreasure {
				tl = append(tl, t...)
			}

			p.HeldTreasure = nil
		}
	}

	// Remove empty tiles from the game.
	var tiles []Tile
	for _, t := range gs.Tiles {
		if t.Type == TileTypeEmpty {
			continue
		}

		tiles = append(tiles, t)
	}
	gs.Tiles = tiles

	// Place the dead players' treasure in stacks of three at the end.
	for _, ts := range stack(tl, 3) {
		ts := ts

		gs.Tiles = append(gs.Tiles, Tile{
			Type:     TileTypeTreasure,
			Treasure: &ts,
		})
	}

	gs.Turn++
	gs.Air = 25
	return nil
}

func (gs *State) String() string {
	pm := make(map[int]string)
	for i, p := range gs.Players {
		if _, ok := pm[p.Position]; ok {
			pm[p.Position] = "*"
		} else {
			pm[p.Position] = fmt.Sprint(i)
		}
	}

	str := "------------" + fmt.Sprintf("ROUND %d", gs.Turn) + "------------\n"
	for i, p := range gs.Players {
		str += fmt.Sprintf("\t player %d: held(%d), stashed(%d)\n", i,
			len(p.HeldTreasure), len(p.StashedTreasure))
	}
	str += "\t"
	for i, t := range gs.Tiles {
		if s, ok := pm[i]; ok {
			str += s
			continue
		}

		switch t.Type {
		case TileTypeEmpty:
			str += "."

		case TileTypeTreasure:
			str += "$"

		case TileTypeSubmarine:
			str += "@"
		}
	}
	str += "\n-------------------------------"

	return str
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

func stack(tl []Treasure, size int) []TreasureStack {
	var cur TreasureStack
	var tsl []TreasureStack
	for _, t := range tl {
		cur = append(cur, t)
		if len(cur) >= size {
			tsl = append(tsl, cur)
			cur = nil
		}
	}

	if len(cur) > 0 {
		tsl = append(tsl, cur)
	}

	return tsl
}
