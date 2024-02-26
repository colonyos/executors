package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type StatusMessage struct {
	Status string `json:"status"`
}

type LogMessage struct {
	Log string `json:"log"`
	EOF bool   `json:"eof"`
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

func (handler *DockerHandler) PullImage(image string, logChan chan LogMessage) error {
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
				logChan <- LogMessage{Log: message.Status + "", EOF: true}
				return nil
			} else {
				log.WithFields(log.Fields{"Error": err}).Error("Error decoding JSON")
				return err
			}
		}

		if message.Status != "" {
			logChan <- LogMessage{Log: message.Status + "", EOF: false}
		}

		fmt.Println(message.Status)
	}
}

func (handler *DockerHandler) GetContainerLogsNoTTY(containerID string, logsChan chan LogMessage, errChan chan error) error {
	ctx := context.Background()
	options := container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true, Tail: "all"}

	out, err := handler.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		log.WithFields(log.Fields{"ContainerID": containerID, "Error": err}).Error("Error getting container logs")
		return err
	}
	defer out.Close()

	for {
		header := make([]byte, 8) // Header is 8 bytes
		_, err := io.ReadFull(out, header)
		if err != nil {
			if err == io.EOF {
				// End of the stream, send final log message and break the loop
				logsChan <- LogMessage{Log: "", EOF: true}
				break
			}
			// For non-EOF errors, consider retrying or handling them differently
			log.WithFields(log.Fields{"ContainerID": containerID, "Error": err}).Error("Error reading log header")
			continue // Consider adding a retry mechanism or a more sophisticated error handling approach here
		}

		size := binary.BigEndian.Uint32(header[4:]) // Get the size of the frame
		if size == 0 {
			continue // Skip zero-length frames
		}

		frame := make([]byte, size)
		_, err = io.ReadFull(out, frame)
		if err != nil {
			log.WithFields(log.Fields{"ContainerID": containerID, "Error": err}).Error("Error reading log frame")
			continue // Consider adding a retry mechanism or a more sophisticated error handling approach here
		}

		logMessage := string(frame)                         // Convert the frame to a string
		logsChan <- LogMessage{Log: logMessage, EOF: false} // Send the log message
	}

	return nil
}

func (handler *DockerHandler) GetContainerLogs(containerID string, logsChan chan LogMessage, errChan chan error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	options := container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true, Tail: "all"}
	out, err := handler.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		log.WithFields(log.Fields{"ContainerID": containerID, "Error": err}).Error("Error getting container logs")
		return err
	}
	defer out.Close()

	reader := bufio.NewReader(out)
	var buffer bytes.Buffer
	var lastChar byte

	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				// Send any remaining content in the buffer as a log message
				if buffer.Len() > 0 {
					logsChan <- LogMessage{Log: buffer.String(), EOF: false}
				}
				break
			} else {
				errChan <- err
				return err
			}
		}

		// Handle carriage returns by sending the buffer content before overwriting it
		if b == '\r' {
			if buffer.Len() > 0 && lastChar != '\r' { // Avoid sending empty messages or duplicates
				logsChan <- LogMessage{Log: buffer.String(), EOF: false}
				buffer.Reset()
			}
			// Do not reset the buffer yet, as we might get more characters to overwrite it
		} else if b == '\n' {
			if lastChar == '\r' { // If the newline follows a carriage return, clear the buffer (line overwrite)
				buffer.Reset()
			} else { // Otherwise, it's a new line, send the buffer content
				logsChan <- LogMessage{Log: buffer.String(), EOF: false}
				buffer.Reset()
			}
		} else {
			buffer.WriteByte(b) // Add non-control characters to the buffer
		}

		lastChar = b // Remember the last character for handling \r\n sequences
	}

	// Indicate EOF with an empty log message
	logsChan <- LogMessage{Log: "", EOF: true}

	return nil
}

func (handler *DockerHandler) GetContainerLogsWorking(containerID string, logsChan chan LogMessage, errChan chan error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	options := container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true, Tail: "all"}
	out, err := handler.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		log.WithFields(log.Fields{"ContainerID": containerID, "Error": err}).Error("Error getting container logs")
		return err
	}
	defer out.Close()

	reader := bufio.NewReader(out)
	var buffer bytes.Buffer
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				// Send any remaining content in the buffer as a log message
				if buffer.Len() > 0 {
					logsChan <- LogMessage{Log: buffer.String(), EOF: false}
				}
				break
			} else {
				errChan <- err
				return err
			}
		}

		if b == '\n' || b == '\r' {
			// Send the current content of the buffer as a log message
			fmt.Println("sending log message:", len(buffer.String()))
			logsChan <- LogMessage{Log: buffer.String(), EOF: false}
			buffer.Reset() // Clear the buffer for the next message
			if b == '\n' {
				continue // Skip the rest of the loop for newline characters
			}
		} else {
			buffer.WriteByte(b) // Add the byte to the buffer if it's not a newline or carriage return
		}
	}

	// Indicate EOF with an empty log message
	logsChan <- LogMessage{Log: "", EOF: true}

	return nil
}

func (handler *DockerHandler) GetContainerLogsAlmostWorking(containerID string, logsChan chan LogMessage, errChan chan error) error {
	ctx := context.Background()

	options := container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true, Tail: "all"}
	out, err := handler.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		log.WithFields(log.Fields{"ContainerID": containerID, "Error": err}).Error("Error getting container logs")
		return err
	}
	defer out.Close()

	reader := bufio.NewReader(out)
	var output []byte
	for {
		fmt.Println("Reading")
		b, err := reader.ReadByte()
		fmt.Println("Got")
		if err != nil {
			if err != io.EOF {
				errChan <- err
			}
			break
		}

		if b == '\n' || b == '\r' {
			if len(output) > 0 {
				logsChan <- LogMessage{Log: string(output), EOF: false}
				output = nil // Reset output buffer
			}
			if b == '\n' {
				continue // For newline, wait for the next line
			}
		} else {
			output = append(output, b)
		}
	}

	// Send any remaining output
	if len(output) > 0 {
		logsChan <- LogMessage{Log: string(output), EOF: false}
	}

	// Indicate EOF
	logsChan <- LogMessage{Log: "", EOF: true}

	return nil
}

func (handler *DockerHandler) StartContainer(image string, cmd string, args []string, env map[string]string, processID string, cfsMount string) (string, error) {
	ctx := context.Background()

	volumeBindings := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: cfsMount,
			Target: "/cfs",
		},
	}

	cmdFlat := cmd
	for _, arg := range args {
		cmdFlat += " " + arg
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, "sh")
	cmdArgs = append(cmdArgs, "-c")
	cmdArgs = append(cmdArgs, "mkdir -p /cfs;"+cmdFlat)

	var envArray []string
	for k, v := range env {
		envArray = append(envArray, k+"="+v)
	}

	envArray = append(envArray, "COLONIES_PROCESS_ID="+processID)
	envArray = append(envArray, "PYTHONUNBUFFERED=1")
	envArray = append(envArray, "PYTHONIOENCODING=UTF-8")

	// currentUser, err := user.Current()
	// if err != nil {
	// 	logrus.Fatal(err)
	// }
	// uid, err := strconv.Atoi(currentUser.Uid)
	// if err != nil {
	// 	logrus.Fatal(err)
	// }
	// gid, err := strconv.Atoi(currentUser.Gid)
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	resp, err := handler.cli.ContainerCreate(ctx, &container.Config{
		Image:        image,
		Cmd:          cmdArgs,
		Env:          envArray,
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		//User:  fmt.Sprintf("%d:%d", uid, gid),
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
