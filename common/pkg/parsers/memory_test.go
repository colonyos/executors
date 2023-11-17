package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMemory(t *testing.T) {
	mbStr, err := ParseMemory("1000Mi")
	assert.Nil(t, err)
	assert.Equal(t, mbStr, "1049M")

	mbStr, err = ParseMemory("1000Mib")
	assert.NotNil(t, err) // Errors, Mib suffix is not allowed

	mbStr, err = ParseMemory("1Mi")
	assert.Nil(t, err)
	assert.Equal(t, mbStr, "1M")

	mbStr, err = ParseMemory("1Gi")
	assert.Nil(t, err)
	assert.Equal(t, mbStr, "1074M")

	mbStr, err = ParseMemory("0Gi")
	assert.Nil(t, err)
	assert.Equal(t, mbStr, "0M")

	mbStr, err = ParseMemory("-1Gi")
	assert.NotNil(t, err) // Memory cannot be negative
}

func TestValidateMemory(t *testing.T) {
	err := ValidateMemory("-1Gi")
	assert.NotNil(t, err) // Memory cannot be negative

	err = ValidateMemory("-1")
	assert.NotNil(t, err) // Invalid format

	err = ValidateMemory("")
	assert.NotNil(t, err) // Invalid format

	err = ValidateMemory("1000Mib")
	assert.NotNil(t, err) // Errors, Mib suffix is not allowed
}
