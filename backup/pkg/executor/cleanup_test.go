package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanup(t *testing.T) {
	files := []string{"backups/backup_1685885186.bak", "backups/backup_1685886083.bak", "backups/backup_1685886210.bak", "backups/backup_1685886944.bak", "backups/backup_1685887245.bak", "backups/backup_1685887312.bak", "backups/backup_1685887415.bak", "backups/backup_1685887467.bak", "backups/backup_1685887536.bak", "backups/backup_1685889598.bak", "backups/backup_1685889638.bak", "backups/backup_1685889674.bak", "backups/backup_1685889712.bak", "backups/backup_1685889774.bak"}

	filesToDelete := cleanup(3, files)

	counter := 0
	for _, file := range files {
		if file == "backups/backup_1685889774.bak" {
			counter++
		}
		if file == "backups/backup_1685889712.bak" {
			counter++
		}
		if file == "backups/backup_1685889674.bak" {
			counter++
		}
	}
	assert.Equal(t, counter, 3)

	counter = 0
	for _, file := range filesToDelete {
		if file == "backups/backup_1685889774.bak" {
			counter++
		}
		if file == "backups/backup_1685889712.bak" {
			counter++
		}
		if file == "backups/backup_1685889674.bak" {
			counter++
		}
	}
	assert.Equal(t, counter, 0)
	assert.Len(t, filesToDelete, 11)
	assert.Len(t, files, 14)

	files = []string{"backups/backup_1685889774.bak", "backups/backup_1685885186.bak", "backups/backup_1685886083.bak", "backups/backup_1685886210.bak", "backups/backup_1685886944.bak", "backups/backup_1685887245.bak", "backups/backup_1685887312.bak", "backups/backup_1685887415.bak", "backups/backup_1685887467.bak", "backups/backup_1685887536.bak", "backups/backup_1685889598.bak", "backups/backup_1685889638.bak", "backups/backup_1685889674.bak", "backups/backup_1685889712.bak"}

	filesToDelete = cleanup(3, files)
	counter = 0
	for _, file := range filesToDelete {
		if file == "backups/backup_1685889774.bak" {
			counter++
		}
		if file == "backups/backup_1685889712.bak" {
			counter++
		}
		if file == "backups/backup_1685889674.bak" {
			counter++
		}
	}
	assert.Equal(t, counter, 0)
	assert.Len(t, filesToDelete, 11)
	assert.Len(t, files, 14)

	files = []string{"backups/backup_1685889774.bak", "backups/backup_1685885186.bak"}
	filesToDelete = cleanup(3, files)
	assert.Len(t, filesToDelete, 0)

	files = []string{"backups/backup_1685889774.bak", "backups/backup_1685885186.bak", "backups/backup_1685887467.bak"}
	filesToDelete = cleanup(3, files)
	assert.Len(t, filesToDelete, 0)

	files = []string{"backups/backup_1685889774.bak", "backups/backup_1685885186.bak", "backups/backup_1685887467.bak", "backups/backup_1685886083.bak"}
	filesToDelete = cleanup(3, files)
	assert.Len(t, filesToDelete, 1)
}
