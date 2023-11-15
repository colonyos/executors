#!/bin/bash
#SBATCH --job-name=my_simple_job
#SBATCH --partition=boost_usr_prod
#SBATCH --account=euhpc_d02_030
#SBATCH --nodes=1
#SBATCH --ntasks-per-node=1
#SBATCH --mem=10G
#SBATCH --time=00:05:00
#SBATCH --gres=gpu:1

module load singularity/3.8.7
srun singularity exec --nv ml.sif nvidia-smi
