package slurm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/executors/common/pkg/singularity"
	"github.com/stretchr/testify/assert"
)

func TestSlurmGenerateBatchScript(t *testing.T) {
	partition := "boost_usr_prod"
	account := "euhpc_d02_030"
	module := "singularity/3.8.7"
	nodes := 10
	tasksPerNode := 1
	cpusPerTask := "1000m"
	walltime := 70
	mem := "10Gi"
	gpus := 2
	command := "nvidia-smi"
	image := "/ml.sif"
	processID := core.GenerateRandomID()
	containerFSDir := "/cfs"
	coloniesTLS := "true"
	coloniesServerHost := "colonies.colonyos.io"
	coloniesServerPort := "443"
	colonyID := "cf3032020e6a94d062ac4c5e8e672068afaa78de2b8c3c0d5316197b27e6beae"
	executorID := "f5fb2739943d5c057b24e3e1626bf4b5f1f4064bc4a73cf3f1d2b72f74125834"
	executorPrvKey := "e96a20824cceb346fa62c7180fac51df2817177ac6021e7bf8709939a2873b06"
	workDir := "/scratch/slurm/test_workdir"
	logDir := "/scratch/slurm/test_logs"
	devMode := true

	slurm := CreateSlurm(workDir, logDir, partition, account, module, true)

	envMap := map[string]string{
		"COLONIES_PROCESS": "ProcessValue",
		"ANOTHER_VAR":      "AnotherValue",
	}

	script, err := slurm.GenerateSlurmScript(
		nodes,
		tasksPerNode,
		cpusPerTask,
		walltime,
		mem,
		gpus,
		command,
		image,
		processID,
		nil,
		containerFSDir,
		coloniesTLS,
		coloniesServerHost,
		coloniesServerPort,
		colonyID,
		executorID,
		executorPrvKey,
		envMap,
		devMode)
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
	cpusPerTask := "1000m"
	mem := "100Mi"
	gpus := 0
	command := "python3 --version"
	processID := core.GenerateRandomID()
	logDir := "/scratch/slurm/logs"
	workDir := "/scratch/slurm/workdir"
	image := "python:3.12-rc-bookworm"
	containerFSDir := "/cfs"
	coloniesTLS := "true"
	coloniesServerHost := "colonies.colonyos.io"
	coloniesServerPort := "443"
	colonyID := "cf3032020e6a94d062ac4c5e8e672068afaa78de2b8c3c0d5316197b27e6beae"
	executorID := "f5fb2739943d5c057b24e3e1626bf4b5f1f4064bc4a73cf3f1d2b72f74125834"
	executorPrvKey := "e96a20824cceb346fa62c7180fac51df2817177ac6021e7bf8709939a2873b06"
	devMode := true

	envMap := map[string]string{
		"COLONIES_PROCESS": "ProcessValue",
		"ANOTHER_VAR":      "AnotherValue",
	}

	singularity := singularity.CreateSingularity("/scratch/slurm/images")
	if !singularity.SifExists(image) {
		logs, err := singularity.Build(image)
		assert.Nil(t, err)
		fmt.Println(logs)
	} else {
		fmt.Println(singularity.Sif(image) + " already exists")
	}

	slurm := CreateSlurm(workDir, logDir, partition, account, module, true)

	script, err := slurm.GenerateSlurmScript(
		nodes,
		tasksPerNode,
		cpusPerTask,
		walltime,
		mem,
		gpus,
		command,
		singularity.Sif(image),
		processID,
		nil,
		containerFSDir,
		coloniesTLS,
		coloniesServerHost,
		coloniesServerPort,
		colonyID,
		executorID,
		executorPrvKey,
		envMap,
		devMode)
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
	cpusPerTask := "1000m"
	mem := "100Mi"
	gpus := 0
	command := "python3 --version"
	processID := core.GenerateRandomID()
	logDir := "/scratch/slurm/logs"
	workDir := "/scratch/slurm/workdir"
	image := "python:3.12-rc-bookworm"
	containerFSDir := "/cfs"
	coloniesTLS := "true"
	coloniesServerHost := "colonies.colonyos.io"
	coloniesServerPort := "443"
	colonyID := "cf3032020e6a94d062ac4c5e8e672068afaa78de2b8c3c0d5316197b27e6beae"
	executorID := "f5fb2739943d5c057b24e3e1626bf4b5f1f4064bc4a73cf3f1d2b72f74125834"
	executorPrvKey := "e96a20824cceb346fa62c7180fac51df2817177ac6021e7bf8709939a2873b06"
	devMode := true

	envMap := map[string]string{
		"COLONIES_PROCESS": "ProcessValue",
		"ANOTHER_VAR":      "AnotherValue",
	}

	singularity := singularity.CreateSingularity("/scratch/slurm/images")
	if !singularity.SifExists(image) {
		logs, err := singularity.Build(image)
		assert.Nil(t, err)
		fmt.Println(logs)
	} else {
		fmt.Println(singularity.Sif(image) + " already exists")
	}

	slurm := CreateSlurm(workDir, logDir, partition, account, module, false)

	script, err := slurm.GenerateSlurmScript(
		nodes,
		tasksPerNode,
		cpusPerTask,
		walltime,
		mem,
		gpus,
		command,
		singularity.Sif(image),
		processID,
		nil,
		containerFSDir,
		coloniesTLS,
		coloniesServerHost,
		coloniesServerPort,
		colonyID,
		executorID,
		executorPrvKey,
		envMap,
		devMode)
	assert.Nil(t, err)
	fmt.Println(script)

	_, err = slurm.Submit(script)
	assert.Nil(t, err)

	logChan := make(chan *Log, 1000)
	jobEndedChan := make(chan *JobEnded, 1000)

	slurm.Monitor(logDir, logChan, jobEndedChan)
	log := <-logChan
	assert.True(t, len(log.Log) > 2)
	<-jobEndedChan
}
