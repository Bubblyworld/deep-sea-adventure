package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecisionValues(t *testing.T) {
	for i := 0; i < 256; i++ {
		d := withValue(decisionRoll, i)
		assert.True(t, d&decisionRoll != 0)
		assert.EqualValues(t, i, d.Value())
	}
}
