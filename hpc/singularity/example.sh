#!/bin/bash

#SBATCH --job-name=dbd2f5b8412a1296f0c588d30f6ca69a7f219c57eaf4d4057906b240c31d4ebc
#SBATCH --nodes=1
#SBATCH --mem=10G
#SBATCH --output=/scratch/slurm/logs/dbd2f5b8412a1296f0c588d30f6ca69a7f219c57eaf4d4057906b240c31d4ebc_%j.log
#SBATCH --error=/scratch/slurm/logs/dbd2f5b8412a1296f0c588d30f6ca69a7f219c57eaf4d4057906b240c31d4ebc_%j.log
#SBATCH --time=00:01:00
srun singularity exec --bind /scratch/slurm/cfs:/cfs /scratch/slurm/images/python:3.12-rc-bookworm.sif python3 /src/helloworld.py
