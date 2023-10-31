package slurm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestSlurmGenerateBatchScript(t *testing.T) {
	partition := "boost_usr_prod"
	account := "euhpc_d02_030"
	module := "singularity/3.8.7"
	nodes := 10
	tasksPerNode := 1
	walltime := 70
	mem := "10G"
	gpus := 2
	command := "nvidia-smi"
	image := "/ml.sif"
	processID := core.GenerateRandomID()
	logDir := "/scratch/slurm/logs"
	workDir := "/scratch/slurm/workdir"
	containerWorkDir := "/workdir"

	slurm := CreateSlurm(workDir, containerWorkDir, logDir, partition, account, module, true)

	script, err := slurm.GenerateSlurmScript(nodes, tasksPerNode, walltime, mem, gpus, command, image, processID)
	assert.Nil(t, err)
	fmt.Println(script)
}

func TestSlurmSubmit(t *testing.T) {
	partition := ""
	account := ""
	module := ""
	nodes := 1
	tasksPerNode := 1
	walltime := 70
	mem := ""
	gpus := 0
	command := "python3 --version"
	processID := core.GenerateRandomID()
	logDir := "/scratch/slurm/logs"
	workDir := "/scratch/slurm/workdir"
	containerWorkDir := "/workdir"

	image := "python:3.12-rc-bookworm"
	singularity := CreateSingularity("/scratch/slurm/images")
	if !singularity.SifExists(image) {
		logs, err := singularity.Build(image)
		assert.Nil(t, err)
		fmt.Println(logs)
	} else {
		fmt.Println(singularity.Sif(image) + " already exists")
	}

	slurm := CreateSlurm(workDir, containerWorkDir, logDir, partition, account, module, false)

	script, err := slurm.GenerateSlurmScript(nodes, tasksPerNode, walltime, mem, gpus, command, singularity.Sif(image), processID)
	assert.Nil(t, err)
	fmt.Println(script)

	jobID, err := slurm.Submit(script)
	assert.Nil(t, err)
	assert.True(t, jobID > 0)

	logFilePath := slurm.GetLogFilePath(logDir, processID, jobID)
	fmt.Println(logFilePath)

	logChan := make(chan *Log)
	jobEndedChan := make(chan *JobEnded, 1)
	errChan := make(chan error)

	err = slurm.MonitorExecutionProgress(logFilePath, logChan, jobEndedChan, errChan, true)
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
	partition := ""
	account := ""
	module := ""
	nodes := 1
	tasksPerNode := 1
	walltime := 70
	mem := ""
	gpus := 0
	command := "hostname"
	processID := core.GenerateRandomID()
	logDir := "/scratch/slurm/logs"
	workDir := "/scratch/slurm/workdir"
	containerWorkDir := "/workdir"

	image := "python:3.12-rc-bookworm"
	singularity := CreateSingularity("/scratch/slurm/images")
	if !singularity.SifExists(image) {
		logs, err := singularity.Build(image)
		assert.Nil(t, err)
		fmt.Println(logs)
	} else {
		fmt.Println(singularity.Sif(image) + " already exists")
	}

	slurm := CreateSlurm(workDir, containerWorkDir, logDir, partition, account, module, false)

	script, err := slurm.GenerateSlurmScript(nodes, tasksPerNode, walltime, mem, gpus, command, singularity.Sif(image), processID)
	assert.Nil(t, err)
	_, err = slurm.Submit(script)
	assert.Nil(t, err)

	logChan := make(chan *Log, 1000)
	jobEndedChan := make(chan *JobEnded, 1000)

	slurm.Monitor(logDir, logChan, jobEndedChan)
	log := <-logChan
	assert.True(t, len(log.Log) > 2)
	<-jobEndedChan
}
