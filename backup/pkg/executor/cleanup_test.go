package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanup(t *testing.T) {
	files := []string{"backups/backup_1685885186.tar.gz", "backups/backup_1685886083.tar.gz", "backups/backup_1685886210.tar.gz", "backups/backup_1685886944.tar.gz", "backups/backup_1685887245.tar.gz", "backups/backup_1685887312.tar.gz", "backups/backup_1685887415.tar.gz", "backups/backup_1685887467.tar.gz", "backups/backup_1685887536.tar.gz", "backups/backup_1685889598.tar.gz", "backups/backup_1685889638.tar.gz", "backups/backup_1685889674.tar.gz", "backups/backup_1685889712.tar.gz", "backups/backup_1685889774.tar.gz"}

	filesToDelete := cleanup(3, files)

	counter := 0
	for _, file := range files {
		if file == "backups/backup_1685889774.tar.gz" {
			counter++
		}
		if file == "backups/backup_1685889712.tar.gz" {
			counter++
		}
		if file == "backups/backup_1685889674.tar.gz" {
			counter++
		}
	}
	assert.Equal(t, counter, 3)

	counter = 0
	for _, file := range filesToDelete {
		if file == "backups/backup_1685889774.tar.gz" {
			counter++
		}
		if file == "backups/backup_1685889712.tar.gz" {
			counter++
		}
		if file == "backups/backup_1685889674.tar.gz" {
			counter++
		}
	}
	assert.Equal(t, counter, 0)
	assert.Len(t, filesToDelete, 11)
	assert.Len(t, files, 14)

	files = []string{"backups/backup_1685889774.tar.gz", "backups/backup_1685885186.tar.gz", "backups/backup_1685886083.tar.gz", "backups/backup_1685886210.tar.gz", "backups/backup_1685886944.tar.gz", "backups/backup_1685887245.tar.gz", "backups/backup_1685887312.tar.gz", "backups/backup_1685887415.tar.gz", "backups/backup_1685887467.tar.gz", "backups/backup_1685887536.tar.gz", "backups/backup_1685889598.tar.gz", "backups/backup_1685889638.tar.gz", "backups/backup_1685889674.tar.gz", "backups/backup_1685889712.tar.gz"}

	filesToDelete = cleanup(3, files)
	counter = 0
	for _, file := range filesToDelete {
		if file == "backups/backup_1685889774.tar.gz" {
			counter++
		}
		if file == "backups/backup_1685889712.tar.gz" {
			counter++
		}
		if file == "backups/backup_1685889674.tar.gz" {
			counter++
		}
	}
	assert.Equal(t, counter, 0)
	assert.Len(t, filesToDelete, 11)
	assert.Len(t, files, 14)

	files = []string{"backups/backup_1685889774.tar.gz", "backups/backup_1685885186.tar.gz"}
	filesToDelete = cleanup(3, files)
	assert.Len(t, filesToDelete, 0)

	files = []string{"backups/backup_1685889774.tar.gz", "backups/backup_1685885186.tar.gz", "backups/backup_1685887467.tar.gz"}
	filesToDelete = cleanup(3, files)
	assert.Len(t, filesToDelete, 0)

	files = []string{"backups/backup_1685889774.tar.gz", "backups/backup_1685885186.tar.gz", "backups/backup_1685887467.tar.gz", "backups/backup_1685886083.tar.gz"}
	filesToDelete = cleanup(3, files)
	assert.Len(t, filesToDelete, 1)
}
