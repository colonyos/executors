package executor

const SlurmBatchTemplate = `#!/bin/bash

#SBATCH --job-name={{.JobName}}
{{- if .Partition}}
#SBATCH --partition={{.Partition}}
{{- end}}
{{- if .Account}}
#SBATCH --account={{.Account}}
{{- end}}
{{- if .Nodes}}
#SBATCH --nodes={{.Nodes}}
{{- end}}
{{- if .TasksPerNode}}
#SBATCH --ntasks-per-node={{.TasksPerNode}}
{{- end}}
{{- if .Memory}}
#SBATCH --mem={{.Memory}}
{{- end}}
{{- if gt .GPUs 0}}
{{- if .GRES}}
#SBATCH --gres=gpu:{{.GPUs}}
{{- end}}
{{- end}}
#SBATCH --output={{.LogDir}}/{{.ProcessID}}_%j.log
#SBATCH --error={{.LogDir}}/{{.ProcessID}}_%j.log
{{- if .Time}}
#SBATCH --time={{.Time}}
{{- end}}

{{- if .Module}}
module load {{.Module}}
{{- end}}

{{- if .Image}}
{{- if gt .GPUs 0}}
srun singularity exec --nv --bind {{.Bind}} {{.Image}} {{.Command}}
{{- else}}
srun singularity exec --bind {{.Bind}} {{.Image}} {{.Command}}
{{- end}}
{{- else}}
srun {{.Command}}
{{- end}}
`
