// Package state contains types representing an in-game board state and
// interfaces/functions for manipulating it. The exported types provide a
// friendly API for consuming state data, along with various implementations
// of the underlying data structures.
package state

// TreasureType is the kind of treasure a token represents. Treasures are
// marked by their number of spots in the actual game.
type TreasureType int

const (
	TreasureTypeOne      TreasureType = 1
	TreasureTypeTwo      TreasureType = 2
	TreasureTypeThree    TreasureType = 3
	TreasureTypeFour     TreasureType = 4
	treasureTypeSentinel TreasureType = 5 // always last
)

// Treasure in an actual treasure token. It consists of a type and its actual
// value to the holding player. In the actual game, treasure values are in the
// range [4n - 4, 4n - 1], where n in the number of spots.
type Treasure struct {
	Type  TreasureType
	Value int
}

// TreasureStack is a stack of treasures occupying a tile. In the actual game
// treasure stacks contain at most 3 treasures.
type TreasureStack []Treasure

// TileType represents the kind of tile a player is occupying.
type TileType int

const (
	TileTypeSubmarine TileType = 1
	TileTypeTreasure  TileType = 2
	TileTypeEmpty     TileType = 3
	tileTypeSentinel  TileType = 4 // always last
)

// Tile represents a tile that can be occupied by a player.
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

// Done returns true if the player has finished their round, i.e. returned to
// the submarine before the air ran out.
func (p *Player) Done() bool {
	return p.Position == 0 && p.TurnedAround
}

// Stage represents what phase the game is in, i.e. which node in the game
// state DFSA is accepting input.
// TODO: ASCII art of the DFSA.
type Stage int

const (
	StageRoll      Stage = 1
	StagePickUp    Stage = 2
	StageDrop      Stage = 3
	StageTurn      Stage = 4
	StageEndOfGame Stage = 5
)

// Decision is a compact representation of an action by the current player,
// i.e. an input to the game state DFSA.
type Decision uint16

func (d Decision) Value() uint16 {
	return uint16(d) >> 8
}

func withValue(d Decision, n int) Decision {
	return (d & 0xFF) + Decision(n<<8)
}

const (
	decisionRoll      Decision = 1 << 0
	decisionPickUpYes Decision = 1 << 1
	decisionPickUpNo  Decision = 1 << 2
	decisionDropYes   Decision = 1 << 3
	decisionDropNo    Decision = 1 << 4
	decisionTurnYes   Decision = 1 << 5
	decisionTurnNo    Decision = 1 << 6
)

func Roll(n int) Decision {
	return withValue(decisionRoll, n)
}

func PickUp() Decision {
	return decisionPickUpYes
}

func DontPickUp() Decision {
	return decisionPickUpNo
}

func Drop(n int) Decision {
	return withValue(decisionDropYes, n)
}

func DontDrop() Decision {
	return decisionDropNo
}

func Turn() Decision {
	return decisionTurnYes
}

func DontTurn() Decision {
	return decisionTurnNo
}

// State represents the current state of the game.
type State interface {
	// Round returns which round we're currently in, starting at 1. There are
	// 3 rounds in a standard game of deep sea adventure.
	Round() int

	// Stage returns the current stage of the game state DFSA. Each round
	// starts in the rolling stage in a standard game.
	Stage() Stage

	// Air is how much air is left in the submarine. Each round starts with
	// 25 air units in the submarine in a standard game.
	Air() int

	// CurrentPlayer returns the index of the player whose turn it is. In a
	// standard game of deep sea adventure, the first round is started by the
	// player who was most recently in the sea. On subsequent rounds the player
	// who was deepest at the end of the previous round goes first.
	CurrentPlayer() int

	// Players returns the list of players participating in the game. There are
	// between 2 and 6 players in a standard game.
	Players() []Player

	// Tiles returns the list of tiles currently on board. Tiles change at the
	// end of each round as players die or make it back to the submarine with
	// treasure.
	Tiles() []Tile

	// ValidDecisions returns a list of decisions that can be made in the
	// current game state.
	ValidDecisions() []Decision

	// Do performs the given decision, mutating the state.
	Do(Decision) error

	// Undo reverses the last decision that was made, mutating the state.
	Undo() error
}
