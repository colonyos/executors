package executor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlurmGenerateBatchScript(t *testing.T) {
	slurm := &Slurm{}
	//slurm.GenerateBatchScript("", "", "singularity/3.8.7", 0, 10)
	script, err := slurm.GenerateBatchScript("", "", "", 1, "10G", 10, "nvidia-smi", "/ml.sif")
	assert.Nil(t, err)
	fmt.Println(script)
}

func TestSlurmSubmit(t *testing.T) {
	slurm := &Slurm{}
	script, err := slurm.GenerateBatchScript("", "", "", 1, "", 0, "hostname", "/home/johan/ml.sif")
	assert.Nil(t, err)
	fmt.Println(script)

	jobID, err := slurm.Submit(script)
	assert.Nil(t, err)
	fmt.Println("jobID:", jobID)

	for {
		status, err := slurm.GetJobStatus(jobID)
		assert.Nil(t, err)
		fmt.Println(status)
		time.Sleep(1 * time.Second)
	}
}
