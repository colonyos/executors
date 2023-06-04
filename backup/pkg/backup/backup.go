package backup

import (
	"os"
	"os/exec"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func ExecBackupDB() (int64, int64, string, string, error) {
	now := time.Now()
	timestamp := time.Now().Unix()
	filename := "backup_" + strconv.FormatInt(timestamp, 10) + ".tar.gz"
	path := "/tmp"
	filepath := path + "/" + filename
	command := "pg_dump -h localhost -p 5432 -U postgres -Ft | gzip > " + filepath
	cmd := exec.Command("sh", "-c", command)

	if err := cmd.Start(); err != nil {
		log.Error(err)
		return -1, 0, "", "", err
	}

	if err := cmd.Wait(); err != nil {
		log.Error(err)
		return -1, 0, "", "", err
	}

	end := time.Now()
	execTime := end.Unix() - now.Unix()

	fi, err := os.Stat(filepath)
	if err != nil {
		log.Error(err)
		return -1, 0, "", "", err
	}
	size := fi.Size()

	return size, execTime, path, filename, nil
}
