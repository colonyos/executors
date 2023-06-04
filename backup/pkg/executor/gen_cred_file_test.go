package executor

import (
	"os"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

func TestCredFile(t *testing.T) {
	err := genCredFile("localhost", 5432, "postgres", "postgres", "rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7")
	assert.Nil(t, err)

	homedir, err := homedir.Dir()
	assert.Nil(t, err)

	cred, err := os.ReadFile(homedir + "/" + ".pgpass")
	assert.Nil(t, err)
	assert.Equal(t, "localhost:5432:postgres:postgres:rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7", string(cred))
}
