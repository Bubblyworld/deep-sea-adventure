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

type GameState struct {
	Air     int
	Round   int
	Players []Player

	// Map is stored as a list of TreasureStacks. Map[0] is always an empty
	// treasure stack and represents the starting submarine. An empty treasure
	// stack that isn't the submarine represents an empty tile that a player
	// can drop treasure stacks onto.
	Map []TreasureStack
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
	res.Map = append(res.Map, nil) // submarine
	for _, tt := range getTreasureTypes() {
		vl := getTreasureValues(tt)
		rand.Shuffle(len(vl), func(i, j int) {
			vl[i], vl[j] = vl[j], vl[i]
		})

		for _, v := range vl {
			res.Map = append(res.Map, TreasureStack{
				Treasure{
					Type:  tt,
					Value: v,
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
