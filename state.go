package main

import "math/rand"

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

type GameState struct {
	Air     int
	Round   int
	Players []Player
	Map     []Tile
}

func NewGameState(players []Player) *GameState {
	res := GameState{
		Air:     25,
		Round:   1,
		Players: players,
	}

	// The initial map is always laid out as follows. The first tile is the
	// starting submarine. The next 8 tiles are the treasures marked with one
	// dot in random order. Then come the two, three and four dot treasures in
	// similar fashion.
	res.Map = append(res.Map, Tile{
		Type: TileTypeSubmarine,
	})

	for _, tt := range getTreasureTypes() {
		vl := getTreasureValues(tt)
		rand.Shuffle(len(vl), func(i, j int) {
			vl[i], vl[j] = vl[j], vl[i]
		})

		for _, v := range vl {
			res.Map = append(res.Map, Tile{
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
