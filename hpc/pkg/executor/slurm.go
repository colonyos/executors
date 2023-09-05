package executor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	PENDING      int = 0  // The job is awaiting resource allocation.
	CONFIGURING      = 1  // The job has been allocated resources, but are waiting for them to become usable
	RUNNING          = 2  // Resources have been allocated to the job and it is currently executing
	COMPLETED        = 3  // The job has finished successfully
	COMPLETING       = 4  // The job has finished executing and is in the process of cleaning up
	SUSPENDED        = 5  // Job execution has been paused, often due to a higher-priority job requiring resources
	CANCELLED        = 6  // The job was explicitly cancelled by the user or system administrator
	FAILED           = 7  // The job terminated before completing
	PREEMPTED        = 8  // The job was terminated because a higher-priority job needed its resources
	REVOKED          = 9  // Resource allocation for the job was revoked due to other higher priority jobs
	SPECIAL_EXIT     = 10 // The job was requeued in a held state
	TIMEOUT          = 11 // The job ran out of time and was killed by the system
	NODE_FAIL        = 12 // The job was terminated because the node on which it was running failed
	OTHER            = 13
)

type Slurm struct {
}

type Log struct {
	Log       string
	ProcessID string
}

type JobEnded struct {
	ProcessID string
	JobID     int
	JobStatus int
}

type JobParams struct {
	Dir       string
	Partition string
	Account   string
	Module    string
	Nodes     int
	Memory    string
	JobName   string
	GPUs      int
	Command   string
	Image     string
	ProcessID string
}

func (slurm *Slurm) GenerateSlurmScript(dir string, partition string, account string, module string, nodes int, mem string, gpus int, command string, image string, processID string) (string, error) {
	os.MkdirAll(dir, 0755)

	tmpl := `#!/bin/bash

#SBATCH --job-name={{.JobName}}
{{- if .Partition}}
#SBATCH --partition={{.Partition}}
{{- end}}
{{- if .Account}}
#SBATCH --account={{.Account}}
{{- end}}
#SBATCH --nodes={{.Nodes}}
{{- if .Memory}}
#SBATCH --mem={{.Memory}}
{{- end}}
{{- if gt .GPUs 0}}
#SBATCH --gres=gpu:{{.GPUs}}
{{- end}}
#SBATCH --output={{.Dir}}/{{.ProcessID}}_%j.log
#SBATCH --error={{.Dir}}/{{.ProcessID}}_%j.log

{{- if .Partition}}
module load singularity/3.8.7
{{- end}}

{{- if .Image}}
{{- if gt .GPUs 0}}
srun singularity exec --nv {{.Image}} {{.Command}}
{{- else}}
srun singularity exec {{.Image}} {{.Command}}
{{- end}}
{{- else}}
srun {{.Command}}
{{- end}}
`
	params := JobParams{
		Dir:       dir,
		Partition: partition,
		Account:   account,
		Module:    module,
		Nodes:     nodes,
		Memory:    mem,
		JobName:   "my_simple_job",
		GPUs:      gpus,
		Command:   command,
		Image:     image,
		ProcessID: processID,
	}

	t := template.Must(template.New("sbatchTemplate").Parse(tmpl))
	var scriptContent bytes.Buffer
	if err := t.Execute(&scriptContent, params); err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Error executing template")
		return "", err
	}
	return scriptContent.String(), nil
}

func (slurm *Slurm) GetLogFilePath(dir string, processID string, jobID int) string {
	return dir + "/" + processID + "_" + strconv.Itoa(jobID) + ".log"
}

func (slurm *Slurm) Submit(script string) (int, error) {
	tmpfile, err := ioutil.TempFile("", "sbatch_script.*.sh")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmpfile.Name())

	fmt.Println(tmpfile.Name())

	_, err = tmpfile.WriteString(script)
	if err != nil {
		return -1, err
	}
	tmpfile.Close()

	cmd := exec.Command("sbatch", tmpfile.Name())
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		fmt.Println(exitError)
		if ok {
			log.WithFields(log.Fields{"Error": exitError, "Stderr": stderr.String(), "ExitCode": exitError.ExitCode()}).Error("Command exited with error")
			// TODO: return a error message
		}
		return 0, err
	}

	output := strings.TrimSpace(out.String())
	parts := strings.Split(output, " ")
	if len(parts) < 4 {
		return 0, fmt.Errorf("Unexpected output format: %s", output)
	}

	jobID, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, fmt.Errorf("Could not parse jobID: %v", err)
	}

	return jobID, nil
}

func (slurm *Slurm) GetJobStatus(jobID int) (int, error) {
	cmdString := fmt.Sprintf("scontrol show job %d | grep -o 'JobState=[^ ]*' | cut -d= -f2", jobID)
	cmd := exec.Command("bash", "-c", cmdString)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return -1, err
	}

	jobStatus := strings.TrimSpace(out.String())
	switch jobStatus {
	case "PENDING":
		return PENDING, nil
	case "CONFIGURING":
		return CONFIGURING, nil
	case "RUNNING":
		return RUNNING, nil
	case "COMPLETED":
		return COMPLETED, nil
	case "COMPLETING":
		return COMPLETING, nil
	case "SUSPENDED":
		return SUSPENDED, nil
	case "CANCELLED":
		return CANCELLED, nil
	case "FAILED":
		return FAILED, nil
	case "PREEMPTED":
		return PREEMPTED, nil
	case "REVOKED":
		return REVOKED, nil
	case "SPECIAL_EXIT":
		return SPECIAL_EXIT, nil
	case "TIMEOUT":
		return TIMEOUT, nil
	case "NODE_FAIL":
		return NODE_FAIL, nil
	default:
		return OTHER, nil
	}
}

type processRecord struct {
	processID string
	errChan   chan error
	replyChan chan bool
}

func (slurm *Slurm) Monitor(dir string, logChan chan *Log, jobEndedChan chan *JobEnded) {
	processes := make(map[string]chan error)
	addProcessChan := make(chan *processRecord)
	deleteProcessChan := make(chan string)
	existsChan := make(chan *processRecord)
	go func() {
		for {
			select {
			case addProcess := <-addProcessChan:
				processes[addProcess.processID] = addProcess.errChan
			case processID := <-deleteProcessChan:
				delete(processes, processID)
			case process := <-existsChan:
				if _, exists := processes[process.processID]; exists {
					process.replyChan <- true
				} else {
					process.replyChan <- false
				}
			}
		}

	}()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			files, err := os.ReadDir(dir)
			if err != nil {
				fmt.Println("Error reading directory:", err)
				time.Sleep(1 * time.Second)
				continue
			}

			for _, file := range files {
				if fileInfo, err := file.Info(); err == nil {
					if !fileInfo.IsDir() {
						processID, _, err := slurm.parseLogPath(file.Name())
						if err != nil {
							fmt.Println("Failed to parse logPath")
							continue
						}
						reply := make(chan bool)
						existsChan <- &processRecord{processID: processID, replyChan: reply}
						exists := <-reply
						if !exists {
							logPath := dir + "/" + file.Name()
							fmt.Println("Monitoring " + logPath)
							errChan := make(chan error)
							addProcessChan <- &processRecord{processID: processID, errChan: errChan}
							err = slurm.MonitorExecutionProgress(logPath, logChan, jobEndedChan, errChan)
							if err != nil {
								log.WithFields(log.Fields{"Error": err}).Error("Failed to monitor slurm job")
							}

							go func(processID string, logPath string) {
								err := <-errChan
								if err != nil {
									log.WithFields(log.Fields{"Error": err}).Error("Error monitoring Slurm execution")
								}
								deleteProcessChan <- processID
								err = os.Remove(logPath)
								if err != nil {
									log.WithFields(log.Fields{"Error": err, "LogPath": logPath}).Error("Failed to remove logfile")
								}
							}(processID, logPath)
						}

					}
				} else {
					fmt.Println("Error getting file info:", err)
					continue
				}

			}
		}
	}()
}

func (slurm *Slurm) parseLogPath(logPath string) (string, int, error) {
	logfileName := filepath.Base(logPath)
	parts := strings.Split(logfileName, "_")

	if len(parts) != 2 {
		return "", -1, errors.New("Failed to parse logPath")
	}

	processID := parts[0]

	parts = strings.Split(parts[1], ".")
	if len(parts) != 2 {
		return "", -1, errors.New("Failed to parse logPath")
	}

	jobID, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", -1, errors.New("Failed to parse jobID in logPath")
	}

	return processID, jobID, nil
}

func (slurm *Slurm) MonitorExecutionProgress(logPath string, logChan chan *Log, jobEndedChan chan *JobEnded, errChan chan error) error {
	processID, jobID, err := slurm.parseLogPath(logPath)
	if err != nil {
		return errors.New("Failed to parse jobID in logPath")
	}

	go func() {
		for {
			file, err := os.Open(logPath)
			if err != nil {
				log.Debug(fmt.Errorf("Error opening file: %w", err))
				time.Sleep(1 * time.Second)
				continue
			}
			defer file.Close()

			initialSize := int64(0)

			reader := bufio.NewReader(file)
			for {
				_, err := file.Seek(initialSize, io.SeekStart)
				if err != nil {
					err := fmt.Errorf("Error seeking to last known size: %w", err)
					log.Error(err)
					errChan <- err
					return
				}

				line, err := reader.ReadString('\n')
				if err != nil && err != io.EOF && err != os.ErrClosed {
					log.Debug(fmt.Errorf("Error reading line: %w", err))
					continue
				}

				if len(line) > 0 {
					logChan <- &Log{ProcessID: processID, Log: line}
				}
				if err != nil && err == io.EOF {
					jobStatus, err := slurm.GetJobStatus(jobID)
					if err != nil {
						log.WithFields(log.Fields{"Error": err}).Error("Error checking job status")
					}
					if jobStatus > RUNNING {
						jobEndedChan <- &JobEnded{ProcessID: processID, JobID: jobID, JobStatus: jobStatus}
						// TODO: make it configurabel to remove logpath
						// err := os.Remove(logPath)
						// if err != nil {
						// 	errChan <- err
						// 	return
						// }
						errChan <- nil
						return // We are done
					}
					time.Sleep(1 * time.Second)
				}

				newSize, err := file.Seek(0, io.SeekEnd)
				if err != nil {
					err := fmt.Errorf("Error updating file size: %w", err)
					log.Error(err)
					errChan <- err
					return
				}
				initialSize = newSize
			}
		}
	}()

	return nil
}
