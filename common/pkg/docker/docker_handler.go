package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type StatusMessage struct {
	Status string `json:"status"`
}

type DockerHandler struct {
	cli *client.Client
}

func CreateDockerHandler() (*DockerHandler, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerHandler{
		cli: cli,
	}, nil
}

func (handler *DockerHandler) PullImage(image string, logChan chan string, doneChan chan string) error {
	ctx := context.Background()

	imageName := image
	out, err := handler.cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		log.WithFields(log.Fields{"Image": imageName, "Error": err}).Error("Error pulling image")
		return err
	}
	defer out.Close()

	decoder := json.NewDecoder(out)

	for {
		var message StatusMessage
		err := decoder.Decode(&message)
		if err != nil {
			if err == io.EOF {
				doneChan <- message.Status
				return nil
			} else {
				log.WithFields(log.Fields{"Error": err}).Error("Error decoding JSON")
				return err
			}
		}
		logChan <- message.Status + "\n"

		log.WithFields(log.Fields{"Image": imageName}).Info(message.Status)
	}
}

// func filterNonPrintableChars(s string) string {
// 	var result []rune
// 	for _, r := range s {
// 		if unicode.IsPrint(r) || unicode.IsSpace(r) {
// 			result = append(result, r)
// 		}
// 	}
// 	return string(result)
// }

func isPrintable(r rune) bool {
	// Defines printable ASCII characters and includes space, tab, and newline
	return (r >= 32 && r <= 126) || r == '\n' || r == '\t'
}

func filterNonPrintableChars(str string) string {
	filtered := ""
	for _, r := range str {
		if isPrintable(r) {
			filtered += string(r)
		}
	}
	return filtered
}

func (handler *DockerHandler) GetContainerLogs(containerID string, logsChan chan string, eofChan chan string, errChan chan error) error {
	ctx := context.Background()

	options := container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true, Tail: "all"}

	out, err := handler.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		log.WithFields(log.Fields{"ContainerID": containerID, "Error": err}).Error("Error getting container logs")
		return err
	}
	defer out.Close()

	buf := make([]byte, 102)
	for {
		n, err := out.Read(buf)
		fmt.Println("n:", n)
		if err != nil {
			if err == io.EOF {
				l := filterNonPrintableChars(string(buf[:n]))
				fmt.Println("l:", len(l))
				log.WithFields(log.Fields{"ContainerID": containerID}).Info(l)
				eofChan <- l
				break
			}
			errChan <- err
		}

		l := filterNonPrintableChars(string(buf[:n]))
		log.WithFields(log.Fields{"ContainerID": containerID}).Info(l)

		logsChan <- l
	}

	return nil
}

func (handler *DockerHandler) StartContainer(image string, cmd string, args []string, cfsMount string) (string, error) {
	ctx := context.Background()

	volumeBindings := []mount.Mount{
		{
			Type: mount.TypeBind,
			//Source: "/scratch/slurm/docker/cfs",
			Source: cfsMount,
			Target: "/cfs",
		},
	}

	cmdArgs := append([]string{cmd}, args...)

	resp, err := handler.cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   cmdArgs,
	}, &container.HostConfig{
		Mounts: volumeBindings,
	}, nil, nil, "")
	if err != nil {
		log.WithFields(log.Fields{"Image": image, "Error": err}).Error("Error creating container")
		return "", err
	}

	if err := handler.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.WithFields(log.Fields{"ContainerID": resp.ID, "Error": err}).Error("Error starting container")
		return "", err
	}

	log.WithFields(log.Fields{"ContainerID": resp.ID}).Info("Container started")

	return resp.ID, nil
}
