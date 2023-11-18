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

export COLONIES_TLS="{{.ColoniesTLS}}"
export COLONIES_SERVER_TLS="{{.ColoniesTLS}}"
export COLONIES_SERVER_HOST="{{.ColoniesServerHost}}"
export COLONIES_SERVER_PORT="{{.ColoniesServerPort}}"
export COLONIES_COLONY_ID="{{.ColonyID}}"
export COLONIES_EXECUTOR_ID="{{.ExecutorID}}"
export COLONIES_EXECUTOR_PRVKEY="{{.ExecutorPrvKey}}"
export COLONIES_PROCESS_ID="{{.ProcessID}}"
export COLONIES_PROCESS="{{.Process}}"

{{- range $key, $value := .EnvMap }} 
export {{ $key }}="{{ $value }}"
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
