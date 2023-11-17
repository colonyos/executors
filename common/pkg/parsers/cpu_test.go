package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCPU(t *testing.T) {
	cpu, err := ParseCPU("10m")
	assert.Nil(t, err)
	assert.Equal(t, cpu, "1")

	cpu, err = ParseCPU("2400m")
	assert.Nil(t, err)
	assert.Equal(t, cpu, "3")

	cpu, err = ParseCPU("2000m")
	assert.Nil(t, err)
	assert.Equal(t, cpu, "2")

	cpu, err = ParseCPU("2000")
	assert.NotNil(t, err)

	cpu, err = ParseCPU("2000M")
	assert.NotNil(t, err)

	cpu, err = ParseCPU("")
	assert.NotNil(t, err)

	cpu, err = ParseCPU("-10000m")
	assert.NotNil(t, err)
}

func TestValidateCPU(t *testing.T) {
	err := ValidateCPU("2000")
	assert.NotNil(t, err)

	err = ValidateCPU("2000M")
	assert.NotNil(t, err)

	err = ValidateCPU("")
	assert.NotNil(t, err)

	err = ValidateCPU("-10000m")
	assert.NotNil(t, err)
}
