package slurm

import (
	"bytes"
	"encoding/base64"
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

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/executors/common/pkg/parsers"
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
	fsDir     string
	logDir    string
	partition string
	account   string
	module    string
	gres      bool
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

type JobStarted struct {
	ProcessID string
	JobID     int
	JobStatus int
	NodeList  []string
}

type JobParams struct {
	LogDir       string
	Partition    string
	Account      string
	Module       string
	Nodes        int
	TasksPerNode int
	CPUsPerTask  int
	Time         string
	Memory       string
	JobName      string
	GPUs         int
	Command      string
	Image        string
	ProcessID    string
	Process      string
	Bind         string
	GRES         bool
	ROCm         bool
	DevMode      bool
	EnvMap       map[string]string
}

func CreateSlurm(fsDir string, logDir string, partition string, account string, module string, gres bool) *Slurm {
	slurm := &Slurm{
		fsDir:     fsDir,
		logDir:    logDir,
		partition: partition,
		account:   account,
		module:    module,
		gres:      gres,
	}

	os.MkdirAll(fsDir, 0755)
	os.MkdirAll(logDir, 0755)

	return slurm
}

func (slurm *Slurm) GenerateSlurmScript(
	nodes int,
	tasksPerNode int,
	cpusPerTask string,
	walltime int,
	mem string,
	gpus int,
	initCommand string,
	command string,
	image string,
	processID string,
	process *core.Process,
	containerFsDir string,
	envMap map[string]string,
	devMode bool,
	rocm bool) (string, error) {

	if initCommand != "" {
		command = initCommand + ";" + command
	}

	var processJSON string
	var err error
	if process != nil {
		processJSON, err = process.ToJSON()
		if err != nil {
			return "", err
		}
	}

	processBase64 := base64.StdEncoding.EncodeToString([]byte(processJSON))

	parsedMem, err := parsers.ParseMemory(mem)
	if err != nil {
		return "", err
	}

	parsedCPUPerTask, err := parsers.ParseCPU(cpusPerTask)
	if err != nil {
		return "", err
	}

	parsedCPUPerTaskInt, err := strconv.Atoi(parsedCPUPerTask)
	if err != nil {
		return "", err
	}

	params := JobParams{
		LogDir:       slurm.logDir,
		Partition:    slurm.partition,
		Account:      slurm.account,
		Module:       slurm.module,
		Nodes:        nodes,
		TasksPerNode: tasksPerNode,
		CPUsPerTask:  parsedCPUPerTaskInt,
		Time:         parsers.ParseWalltime(walltime),
		Memory:       parsedMem,
		JobName:      processID,
		GPUs:         gpus,
		Command:      command,
		Image:        image,
		ProcessID:    processID,
		Process:      processBase64,
		GRES:         slurm.gres,
		ROCm:         rocm,
		Bind:         slurm.fsDir + ":" + containerFsDir,
		DevMode:      devMode,
		EnvMap:       envMap,
	}

	t := template.Must(template.New("sbatchTemplate").Parse(SlurmBatchTemplate))
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

// getNodeNames retrieves the names of the compute nodes for a given job ID using Slurm.
func getNodeNames(jobID int) ([]string, error) {
	jobIDStr := strconv.Itoa(jobID)
	cmd := exec.Command("scontrol", "show", "job", jobIDStr)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute scontrol command: %v", err)
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "NodeList=") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				nodeList := strings.TrimSpace(parts[1])
				if strings.Contains(nodeList, "null") {
					return nil, errors.New("Node names not available yet")
				}
				nodes := parseNodeList(nodeList)
				return nodes, nil
			}
		}
	}

	return nil, fmt.Errorf("node names not found for job ID: %s", jobID)
}

// parseNodeList parses the NodeList string from Slurm into individual node names
func parseNodeList(nodeList string) []string {
	// Slurm may provide node lists in various formats (e.g., "node[1-3]" or "node1,node2,node3")
	// This function assumes a simple comma-separated list for demonstration purposes
	// You may need to implement more complex parsing logic for your specific use case
	if strings.Contains(nodeList, "[") && strings.Contains(nodeList, "]") {
		// Example: "node[1-3]" -> ["node1", "node2", "node3"]
		return expandNodeRange(nodeList)
	}

	// Split by comma
	nodes := strings.Split(nodeList, ",")
	return nodes
}

// expandNodeRange expands a node range string like "node[1-3]" into individual node names
func expandNodeRange(nodeRange string) []string {
	var nodes []string

	// Find the prefix (e.g., "node") and the range (e.g., "1-3")
	prefixEnd := strings.Index(nodeRange, "[")
	rangeStr := nodeRange[prefixEnd+1 : len(nodeRange)-1]
	prefix := nodeRange[:prefixEnd]

	// Split the range (e.g., "1-3" -> ["1", "3"])
	rangeParts := strings.Split(rangeStr, "-")
	if len(rangeParts) != 2 {
		return []string{nodeRange} // Invalid range, return as is
	}

	// Convert range parts to integers
	start, err1 := strconv.Atoi(rangeParts[0])
	end, err2 := strconv.Atoi(rangeParts[1])
	if err1 != nil || err2 != nil {
		return []string{nodeRange} // Invalid range, return as is
	}

	for i := start; i <= end; i++ {
		nodes = append(nodes, fmt.Sprintf("%s%d", prefix, i))
	}

	return nodes
}

func (slurm *Slurm) Submit(script string) (int, error) {
	tmpfile, err := ioutil.TempFile("", "sbatch_script.*.sh")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmpfile.Name())

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
		if ok {
			log.WithFields(log.Fields{"Error": exitError, "Stderr": stderr.String(), "ExitCode": exitError.ExitCode()}).Error("Command exit with error")
			return 0, errors.New(exitError.Error() + ":" + stderr.String())
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

func (slurm *Slurm) Monitor(dir string, logChan chan *Log, jobStartedChan chan *JobStarted, jobEndedChan chan *JobEnded) {
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
							errChan := make(chan error)
							addProcessChan <- &processRecord{processID: processID, errChan: errChan}
							err = slurm.MonitorExecutionProgress(logPath, logChan, jobStartedChan, jobEndedChan, errChan, false)
							if err != nil {
								log.WithFields(log.Fields{"Error": err}).Error("Failed to monitor Slurm job")
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

func (slurm *Slurm) MonitorExecutionProgress(logPath string, logChan chan *Log, jobStartedChan chan *JobStarted, jobEndedChan chan *JobEnded, errChan chan error, deleteLogFile bool) error {
	processID, jobID, err := slurm.parseLogPath(logPath)
	if err != nil {
		return errors.New("Failed to parse jobID in logPath")
	}

	go func(jobID int) {
		hasStarted := false
		waitForJobChan := make(chan bool)
		for {
			file, err := os.Open(logPath)
			if err != nil {
				log.Debug(fmt.Errorf("Error opening file: %w", err))
				time.Sleep(1 * time.Second)
				continue
			}
			defer file.Close()

			pos := int64(0)
			for {
				_, err := file.Seek(pos, io.SeekStart)
				if err != nil {
					err := fmt.Errorf("Error seeking to last known size: %w", err)
					log.Error(err)
					errChan <- err
					return
				}

				content, err := io.ReadAll(file)
				if err != nil && err != io.EOF && err != os.ErrClosed {
					log.Error(fmt.Errorf("Error reading line: %w", err))
					continue
				}

				if len(content) > 0 {
					logChan <- &Log{ProcessID: processID, Log: string(content)}
					pos += int64(len(content))
				} else {
					time.Sleep(1 * time.Second)
					jobStatus, err := slurm.GetJobStatus(jobID)
					if err != nil {
						log.WithFields(log.Fields{"Error": err}).Error("Error checking job status")
					}

					if !hasStarted {
						go func() {
							for {
								nodeList, err := getNodeNames(jobID)
								if err != nil {
									log.Debug(err)
									time.Sleep(1000 * time.Millisecond)
									continue
								}
								hasStarted = true
								jobStartedChan <- &JobStarted{ProcessID: processID, JobID: jobID, JobStatus: jobStatus, NodeList: nodeList}
								waitForJobChan <- true
								break
							}
						}()
					}
					if jobStatus != RUNNING {
						select {
						case <-time.After(10 * time.Second):
							log.WithFields(log.Fields{"JobStatus": jobStatus}).Debug("Failed to get node list")
						case <-waitForJobChan:
						}

						content, err := io.ReadAll(file)
						if err != nil {
							log.Error(fmt.Errorf("Error reading line: %w", err))
						}
						if len(content) > 0 {
							logChan <- &Log{ProcessID: processID, Log: string(content)}
						}

						// TODO: It could happen that the job has ended before jonStartedChan is called

						jobEndedChan <- &JobEnded{ProcessID: processID, JobID: jobID, JobStatus: jobStatus}
						if deleteLogFile {
							err := os.Remove(logPath)
							if err != nil {
								errChan <- err
								return
							}
						}
						errChan <- nil
						return // We are done
					}
				}
				_, err = file.Seek(0, io.SeekEnd)
			}
		}
	}(jobID)

	return nil
}
