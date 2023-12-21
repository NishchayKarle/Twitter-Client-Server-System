#!/bin/bash
#
#SBATCH --mail-user=nishchaykarle@uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj1_benchmark 
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/nishchaykarle/parallel/project-1/project-1-NishchayKarle/proj1/grader
#SBATCH --partition=debug 
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=5:00


module load golang/1.16.2 
go test proj1/twitter -v -count=1
