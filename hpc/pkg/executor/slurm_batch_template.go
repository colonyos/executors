package executor

const SlurmBatchTemplate = `#!/bin/bash

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
#SBATCH --output={{.LogDir}}/{{.ProcessID}}_%j.log
#SBATCH --error={{.LogDir}}/{{.ProcessID}}_%j.log

{{- if .Partition}}
module load singularity/3.8.7
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
