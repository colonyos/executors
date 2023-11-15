#!/bin/bash

#SBATCH --job-name=64c00fa11a176eb41e5d3049efea6739c9db6c41e84e4f39e8b54b3c268f63bc
#SBATCH --nodes=1
#SBATCH --ntasks-per-node=1
#SBATCH --output=/scratch/slurm/logs/64c00fa11a176eb41e5d3049efea6739c9db6c41e84e4f39e8b54b3c268f63bc_%j.log
#SBATCH --error=/scratch/slurm/logs/64c00fa11a176eb41e5d3049efea6739c9db6c41e84e4f39e8b54b3c268f63bc_%j.log
#SBATCH --time=00:01:00
srun singularity exec --nv --bind /scratch/slurm/cfs:/cfs /scratch/slurm/images/tensorflow_tensorflow:2.14.0rc1-gpu.sif nvidia-smi
