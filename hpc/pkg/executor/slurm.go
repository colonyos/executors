package executor

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Slurm struct {
}

type JobParams struct {
	Partition string
	Account   string
	Module    string
	Nodes     int
	Memory    string
	JobName   string
	GPUs      int
	Command   string
	Image     string
}

func (slurm *Slurm) GenerateBatchScript(partition string, account string, module string, nodes int, mem string, gpus int, command string, image string) (string, error) {
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

{{- if .Partition}}
module load singularity/3.8.7
{{- end}}

{{- if gt .GPUs 0}}
srun singularity exec --nv {{.Image}} {{.Command}}
{{- else}}
srun singularity exec {{.Image}} {{.Command}}

{{- end}}
`

	// Populate the template with actual values
	params := JobParams{
		Partition: partition,
		Account:   account,
		Module:    module,
		Nodes:     nodes,
		Memory:    mem,
		JobName:   "my_simple_job",
		GPUs:      gpus,
		Command:   command,
		Image:     image,
	}

	t := template.Must(template.New("sbatchTemplate").Parse(tmpl))
	var scriptContent bytes.Buffer
	if err := t.Execute(&scriptContent, params); err != nil {
		fmt.Println("Error executing template:", err)
		return "", err
	}
	return scriptContent.String(), nil
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
		return 0, err
	}
	tmpfile.Close()

	cmd := exec.Command("sbatch", tmpfile.Name())
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("--------------")
		fmt.Println(err)
		exitError, ok := err.(*exec.ExitError)
		if ok {
			fmt.Printf("Command exited with error: %v\n", exitError)
			fmt.Printf("Stderr: %s\n", stderr.String())
			fmt.Printf("Exit Code: %d\n", exitError.ExitCode())
		}
		fmt.Println(err)
		return 0, err
	}

	output := strings.TrimSpace(out.String())
	parts := strings.Split(output, " ")
	if len(parts) < 4 {
		return 0, fmt.Errorf("unexpected output format: %s", output)
	}

	jobID, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, fmt.Errorf("could not parse job ID: %v", err)
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

	fmt.Println(strings.TrimSpace(out.String()))

	return -1, nil
}
