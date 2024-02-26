package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerHandler(t *testing.T) {
	image := "ubuntu:jammy-20240125"
	cmd := "echo"
	args := []string{"hello world"}
	cfsMount := "/tmp"

	logsChan := make(chan LogMessage, 100)
	errChan := make(chan error, 100)

	handler, err := CreateDockerHandler()
	assert.Nil(t, err)

	err = handler.PullImage(image, logsChan)
	assert.Nil(t, err)

	env := make(map[string]string, 0)

	containerID, err := handler.StartContainer(image, cmd, args, env, "", cfsMount)
	assert.Nil(t, err)

	go func() {
		err := handler.GetContainerLogs(containerID, logsChan, errChan)
		assert.Nil(t, err)
	}()

	for {
		select {
		case err := <-errChan:
			fmt.Println(err)
			assert.Nil(t, err)
			return
		case log := <-logsChan:
			fmt.Println(log.Log)
			assert.NotEmpty(t, log)
			if log.EOF {
				return
			}
		}
	}
}
