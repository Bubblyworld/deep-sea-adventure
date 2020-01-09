package main

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

// KnownGameState
type KnownGameState struct {
}
