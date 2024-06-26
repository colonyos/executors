package slurm

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
{{- if .ROCm}}
#SBATCH --gpus-per-node={{.GPUs}}
{{- end}}
{{- if .TasksPerNode}}
#SBATCH --ntasks-per-node={{.TasksPerNode}}
{{- end}}
{{- if not .DevMode}}
{{- if .CPUsPerTask}}
#SBATCH --cpus-per-task={{.CPUsPerTask}}
{{- end}}
{{- end}}
{{- if not .DevMode}}
{{- if .Memory}}
#SBATCH --mem={{.Memory}}
{{- end}}
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

for varname in $(compgen -v | grep ^COLONIES_); do
    unset $varname
done

for varname in $(compgen -v | grep ^EXECUTOR_); do
    unset $varname
done

for varname in $(compgen -v | grep ^AWS_S3_); do
    unset $varname
done

export COLONIES_PROCESS_ID="{{.ProcessID}}"

{{- range $key, $value := .EnvMap }} 
export {{ $key }}="{{ $value }}"
{{- end}}

{{- if .Image}}
{{- if gt .GPUs 0}}
{{- if .ROCm}}
    srun singularity exec --rocm --bind {{.Bind}} {{.Image}} sh -c "{{.Command}}"
{{- else}}
    srun singularity exec --nv --bind {{.Bind}} {{.Image}} sh -c "{{.Command}}"
{{- end}}
{{- else}}
srun singularity exec --bind {{.Bind}} {{.Image}} sh -c "{{.Command}}"
{{- end}}
{{- else}}
srun {{.Command}}
{{- end}}
`
