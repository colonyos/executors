package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseWalltime(t *testing.T) {
	walltime := ParseWalltime(10)
	assert.Equal(t, walltime, "00:00:10")
}
