package state

import (
	"errors"
	"fmt"
	"math/rand"
)

// standardState is a direct, inefficient implementation of the state
// interface. It has no optimisations, and was built for a quick POC
// integration test.
type standardState struct {
	air       int
	round     int
	stage     Stage
	curPlayer int
	players   []Player
	tiles     []Tile
	history   []*standardState
}

func NewStandardState(players int) *standardState {
	ss := standardState{
		air:       25,
		round:     1,
		stage:     StageRoll,
		curPlayer: 0,
	}

	for i := 0; i < players; i++ {
		ss.players = append(ss.players, Player{})
	}

	// The initial map is always laid out as follows. The first tile is the
	// starting submarine. The next 8 tiles are the treasures marked with one
	// dot in random order. Then come the two, three and four dot treasures in
	// similar fashion.
	ss.tiles = append(ss.tiles, Tile{
		Type: TileTypeSubmarine,
	})

	for _, tt := range getTreasureTypes() {
		vl := getTreasureValues(tt)
		rand.Shuffle(len(vl), func(i, j int) {
			vl[i], vl[j] = vl[j], vl[i]
		})

		for _, v := range vl {
			ss.tiles = append(ss.tiles, Tile{
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

	return &ss
}

func (ss *standardState) Round() int {
	return ss.round
}

func (ss *standardState) Stage() Stage {
	return ss.stage
}

func (ss *standardState) Air() int {
	return ss.air
}

func (ss *standardState) CurrentPlayer() int {
	return ss.curPlayer
}

func (ss *standardState) Players() []Player {
	return ss.players
}

func (ss *standardState) Tiles() []Tile {
	return ss.tiles
}

func (ss *standardState) ValidDecisions() []Decision {
	switch ss.stage {
	case StageRoll:
		var rolls []Decision
		for i := 2; i <= 6; i++ {
			rolls = append(rolls, Roll(i))
		}

		return rolls

	case StagePickUp:
		return []Decision{PickUp(true), PickUp(false)}

	case StageDrop:
		drops := []Decision{Drop(0, false)}
		for i := range ss.players[ss.curPlayer].HeldTreasure {
			drops = append(drops, Drop(i, true))
		}

		return drops

	case StageTurn:
		// If we're at the end of the board we have to turn around.
		if ss.players[ss.curPlayer].Position == len(ss.tiles)-1 {
			return []Decision{Turn(true)}
		}

		return []Decision{Turn(true), Turn(false)}

	case StageEndOfGame:
		return nil
	}

	// Should never be reached.
	panic("unknown game stage in standardState")
}

func (ss *standardState) Do(d Decision) error {
	// We're about to alter the state in some way, so push a copy of the state
	// onto the history stack for Undo() calls.
	ssCopy := ss.clone()
	ss.history = append(ss.history, ssCopy)

	var valid bool
	for _, vd := range ss.ValidDecisions() {
		if vd == d {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("attempted to do invalid decision")
	}

	cp := &ss.players[ss.curPlayer]
	switch ss.stage {
	case StageRoll:
		moves := int(d.Value()) - len(cp.HeldTreasure)
		if moves < 0 {
			moves = 0
		}

		ss.air -= len(cp.HeldTreasure)
		if ss.air < 0 {
			ss.air = 0
		}

		if err := ss.move(cp, moves); err != nil {
			return err
		}

		// If we're standing on a treasure, we may pick it up.
		if ss.tiles[cp.Position].Type == TileTypeTreasure {
			ss.stage = StagePickUp
			return nil
		}

		// If we're standing on an empty square and we have treasure, we can
		// choose to drop one of our treasures.
		if ss.tiles[cp.Position].Type == TileTypeEmpty &&
			len(cp.HeldTreasure) > 0 {

			ss.stage = StageDrop
			return nil
		}

		// If we can't pick up or drop, then we just move on to the next turn.
		return ss.toNextTurn()

	case StagePickUp:
		if d&decisionPickUpYes != 0 {
			if err := ss.pickup(cp); err != nil {
				return err
			}
		}

		return ss.toNextTurn()

	case StageDrop:
		if d&decisionDropYes != 0 {
			if err := ss.drop(cp, int(d.Value())); err != nil {
				return err
			}
		}

		return ss.toNextTurn()

	case StageTurn:
		if d&decisionTurnYes != 0 {
			cp.TurnedAround = true
		}

		ss.stage = StageRoll
		return nil
	}

	// Should never be reached.
	panic("invalid decision for game stage in standardState")
}

func (ss *standardState) Undo() error {
	if len(ss.history) == 0 {
		return errors.New("attempted to undo a state with no history")
	}

	history := ss.history
	*ss = *history[len(history)-1]
	ss.history = history[0 : len(history)-1]

	return nil
}

// toNextTurn performs transition logic to the next player's turn, or possibly
// the end of the game if the third round has ended.
func (ss *standardState) toNextTurn() error {
	// Players who have reached the submarine again no longer need to make any
	// actions. If everyone has reached the submarine, or the oxygen is done,
	// the round is over.
	nextPlayer := (ss.curPlayer + 1) % len(ss.players)
	for nextPlayer != ss.curPlayer {
		if !isFinished(ss.players[nextPlayer]) {
			break // current player is able to take a turn
		}

		nextPlayer = (nextPlayer + 1) % len(ss.players)
	}

	ss.curPlayer = nextPlayer
	cp := ss.players[ss.curPlayer]

	if ss.air <= 0 || isFinished(cp) {
		if err := ss.endRound(); err != nil {
			return err
		}

		if ss.round > 3 {
			ss.stage = StageEndOfGame
		} else {
			ss.stage = StageRoll
		}

		return nil
	}

	// If we're not on the submarine tile and we haven't turned around yet,
	// we have the option of doing so.
	ss.stage = StageRoll
	if cp.Position > 0 && !cp.TurnedAround {
		ss.stage = StageTurn
	}

	return nil
}

// validate returns an error only if the game state is illegal.
func (ss *standardState) validate() error {
	pm := make(map[int]bool)
	for _, p := range ss.players {
		if !ss.inBounds(p.Position) {
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

// pickup causes the player to pick-up the treasure stack at their current
// position. If there isn't a treasure to pick up this will error.
func (ss *standardState) pickup(p *Player) error {
	if err := ss.validate(); err != nil {
		return fmt.Errorf("validation error while picking up: %v", err)
	}

	if ss.tiles[p.Position].Type != TileTypeTreasure {
		return errors.New("player tried to pick up non-treasure tile")
	}

	p.HeldTreasure = append(p.HeldTreasure, *ss.tiles[p.Position].Treasure)
	ss.tiles[p.Position] = Tile{
		Type: TileTypeEmpty,
	}

	return nil
}

// drop causes the player to drop the treasure stack with the given index at
// their current position. If they aren't on an empty tile or the treasure
// stack doesn't exist, this will error.
func (ss *standardState) drop(p *Player, index int) error {
	if err := ss.validate(); err != nil {
		return fmt.Errorf("validation error while dropping: %v", err)
	}

	if index < 0 || index >= len(p.HeldTreasure) {
		return errors.New("player tried to drop non-existent treasure")
	}

	if ss.tiles[p.Position].Type != TileTypeEmpty {
		return errors.New("player tried to drop on non-empty tile")
	}

	ss.tiles[p.Position] = Tile{
		Type:     TileTypeTreasure,
		Treasure: &p.HeldTreasure[index],
	}
	p.HeldTreasure = append(p.HeldTreasure[:index],
		p.HeldTreasure[:index+1]...)

	return nil
}

// move moves the player the given number of spaces, hopping over other
// players as they go. If the player reaches the end of the map, or the
// submarine if they're going backwards, then they stop (no bounceback).
func (ss *standardState) move(p *Player, spaces int) error {
	if err := ss.validate(); err != nil {
		return fmt.Errorf("validation error while moving: %v", err)
	}

	// A map of which squares contain players. Note that we don't have to
	// worry about skipping player 'p' since there's no bounceback.
	sm := make(map[int]bool)
	for _, pl := range ss.players {
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
		if !ss.inBounds(newPos) {
			return nil // already at the end
		}
		for ss.inBounds(newPos) && sm[newPos] {
			newPos += skip // jump over players
		}

		if !ss.inBounds(newPos) || sm[newPos] {
			return nil // can't actually move, too many players in front
		}

		p.Position = newPos
	}

	return nil
}

// endRound kills any players that have yet to reach the submarine and resets
// the state for the next round.
func (ss *standardState) endRound() error {
	if err := ss.validate(); err != nil {
		return fmt.Errorf("validation error while ending turn: %v", err)
	}

	// Reset all the players, keeping treasure if they survived.
	var tl []Treasure
	for i := range ss.players {
		p := &ss.players[i]
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
	for _, t := range ss.tiles {
		if t.Type == TileTypeEmpty {
			continue
		}

		tiles = append(tiles, t)
	}
	ss.tiles = tiles

	// Place the dead players' treasure in stacks of three at the end.
	for _, ts := range stack(tl, 3) {
		ts := ts

		ss.tiles = append(ss.tiles, Tile{
			Type:     TileTypeTreasure,
			Treasure: &ts,
		})
	}

	ss.round++
	ss.air = 25
	ss.curPlayer = 0 // TODO: furthest from submarine
	return nil
}

func (ss *standardState) inBounds(pos int) bool {
	return pos >= 0 && pos < len(ss.tiles)
}

// clone copies every field of the given state by value except the state
// history, which is copied by reference.
func (ss *standardState) clone() *standardState {
	clone := standardState{
		air:       ss.air,
		round:     ss.round,
		stage:     ss.stage,
		curPlayer: ss.curPlayer,
		players:   make([]Player, len(ss.players)),
		tiles:     make([]Tile, len(ss.tiles)),
	}

	for i, p := range ss.players {
		clone.players[i] = p
	}

	for i, t := range ss.tiles {
		clone.tiles[i] = t
	}

	return &clone
}

func isFinished(p Player) bool {
	return p.Position == 0 && p.TurnedAround
}

func getTreasureTypes() []TreasureType {
	var res []TreasureType
	for tt := TreasureTypeOne; tt != treasureTypeSentinel; tt++ {
		res = append(res, tt)
	}

	return res
}

var treasureValues = map[TreasureType][]int{
	TreasureTypeOne:   []int{0, 0, 1, 1, 2, 2, 3, 3},
	TreasureTypeTwo:   []int{4, 4, 5, 5, 6, 6, 7, 7},
	TreasureTypeThree: []int{8, 8, 9, 9, 10, 10, 11, 11},
	TreasureTypeFour:  []int{12, 12, 13, 13, 14, 14, 15, 15},
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

// Compile-time implementation check.
var _ State = (*standardState)(nil)
