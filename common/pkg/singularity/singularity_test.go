package singularity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingularity(t *testing.T) {
	img := "python:3.12-rc-bookworm"
	singularity := CreateSingularity("/scratch/slurm/images")

	singularity.Remove(img)
	exists := singularity.SifExists(img)
	assert.False(t, exists)

	logs, err := singularity.Build(img)
	fmt.Println(logs)

	exists = singularity.SifExists(img)
	assert.True(t, exists)
	assert.Nil(t, err)

	singularity.Remove(img)
	exists = singularity.SifExists(img)
	assert.False(t, exists)
}

func TestSingularitySif(t *testing.T) {
	img := "python:3.12-rc-bookworm"
	singularity := CreateSingularity("/scratch/slurm/images")
	assert.Equal(t, singularity.Sif(img), "/scratch/slurm/images/python:3.12-rc-bookworm.sif")
	img = "tensorflow/tensorflow:2.14.0rc1-gpu"
	assert.Equal(t, singularity.Sif(img), "/scratch/slurm/images/tensorflow_tensorflow:2.14.0rc1-gpu.sif")
	img = ""
	assert.Equal(t, singularity.Sif(img), "/scratch/slurm/images/.sif") // TODO: handle this error
}
