package backup

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackup(t *testing.T) {
	size, execTime, path, filename, err := ExecBackupDB("/tmp")
	assert.Nil(t, err)
	assert.True(t, size > 0)
	assert.True(t, execTime > 0)
	assert.True(t, len(path) > 0)
	assert.True(t, len(filename) > 0)

	err = os.Remove(path + "/" + filename)
	assert.Nil(t, err)
}
