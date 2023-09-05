package executor

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestSlurmGenerateBatchScript(t *testing.T) {
	slurm := &Slurm{}
	dir := "./slurm"
	partition := "boost_usr_prod"
	account := "euhpc_d02_030"
	module := "singularity/3.8.7"
	nodes := 10
	mem := "10G"
	gpus := 2
	command := "nvidia-smi"
	image := "/ml.sif"
	processID := core.GenerateRandomID()
	script, err := slurm.GenerateSlurmScript(dir,
		partition,
		account,
		module,
		nodes,
		mem,
		gpus,
		command,
		image,
		processID)
	assert.Nil(t, err)
	fmt.Println(script)
}

func TestSlurmSubmit(t *testing.T) {
	slurm := &Slurm{}

	dir := "./slurm"
	partition := ""
	account := ""
	module := ""
	nodes := 1
	mem := ""
	gpus := 0
	//command := "hostname"
	command := "python3 --version"
	image := "/home/johan/ml.sif"
	processID := core.GenerateRandomID()
	script, err := slurm.GenerateSlurmScript(dir,
		partition,
		account,
		module,
		nodes,
		mem,
		gpus,
		command,
		image,
		processID)
	assert.Nil(t, err)
	fmt.Println(script)

	jobID, err := slurm.Submit(script)
	assert.Nil(t, err)
	assert.True(t, jobID > 0)

	logFilePath := slurm.GetLogFilePath(dir, processID, jobID)
	fmt.Println(logFilePath)

	logChan := make(chan *Log)
	jobEndedChan := make(chan *JobEnded, 1)
	errChan := make(chan error)

	err = slurm.MonitorExecutionProgress(logFilePath, logChan, jobEndedChan, errChan)
	assert.Nil(t, err)

	log := <-logChan
	fmt.Println(log.ProcessID + ": " + log.Log)
	jobEnded := <-jobEndedChan
	fmt.Println("Process with ID <" + jobEnded.ProcessID + "> running Slurm job with ID <" + strconv.Itoa(jobEnded.JobID) + "> ended with status " + strconv.Itoa(jobEnded.JobStatus))
	assert.Equal(t, jobEnded.JobStatus, COMPLETED)
	err = <-errChan
	assert.Nil(t, err)
}

func TestSlurmMonitor(t *testing.T) {
	slurm := &Slurm{}

	dir := "./slurm"
	partition := ""
	account := ""
	module := ""
	nodes := 1
	mem := ""
	gpus := 0
	command := "hostname"
	//image := "/home/johan/ml.sif"
	image := ""

	for i := 0; i < 100; i++ {
		processID := core.GenerateRandomID()
		script, err := slurm.GenerateSlurmScript(dir,
			partition,
			account,
			module,
			nodes,
			mem,
			gpus,
			command,
			image,
			processID)
		assert.Nil(t, err)
		fmt.Println(script)
		jobID, err := slurm.Submit(script)
		fmt.Println("Submitting Slurm job with ID", jobID)
		assert.Nil(t, err)
	}

	// logChan := make(chan *Log, 1000)
	// jobEndedChan := make(chan *JobEnded, 1000)

	// slurm.Monitor(dir, logChan, jobEndedChan)

	// for {
	// 	select {
	// 	case log := <-logChan:
	// 		fmt.Println(log.ProcessID + ": " + log.Log)
	// 	case jobEnded := <-jobEndedChan:
	// 		fmt.Println("Process with ID <" + jobEnded.ProcessID + "> running Slurm job with ID <" + strconv.Itoa(jobEnded.JobID) + "> ended with status " + strconv.Itoa(jobEnded.JobStatus))
	// 	}
	// }
}
