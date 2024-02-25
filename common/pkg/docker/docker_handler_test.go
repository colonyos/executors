package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerHandler(t *testing.T) {
	image := "ubuntu:latest"
	cmd := "echo"
	args := []string{"hello world"}
	cfsMount := "/tmp"

	logsChan := make(chan string, 100)
	eofChan := make(chan string, 100)
	errChan := make(chan error, 100)
	doneChan := make(chan string, 100)

	handler, err := CreateDockerHandler()
	assert.Nil(t, err)

	err = handler.PullImage(image, logsChan, doneChan)
	assert.Nil(t, err)

	containerID, err := handler.StartContainer(image, cmd, args, cfsMount)
	assert.Nil(t, err)

	go func() {
		err := handler.GetContainerLogs(containerID, logsChan, eofChan, errChan)
		assert.Nil(t, err)
	}()

	for {
		select {
		case log := <-eofChan:
			fmt.Println(log)
			assert.NotEmpty(t, log)
			return
		case err := <-errChan:
			fmt.Println(err)
			assert.Nil(t, err)
			return
		case log := <-logsChan:
			fmt.Println(log)
			assert.NotEmpty(t, log)
		}
	}
}
