#!/bin/bash
#SBATCH --job-name=my_simple_job
#SBATCH --nodes=1
#SBATCH --ntasks-per-node=1
#SBATCH --time=00:05:00

srun singularity exec --nv ml.sif nvidia-smi
