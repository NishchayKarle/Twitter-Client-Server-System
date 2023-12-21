#!/bin/bash
#
#SBATCH --mail-user=nishchaykarle@uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=part6
#SBATCH --output=./out/%j.%N.stdout
#SBATCH --error=./out/%j.%N.stderr
#SBATCH --chdir=/home/nishchaykarle/parallel/project-1/project-1-NishchayKarle/proj1/benchmark
#SBATCH --partition=debug
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=01:00:00


module load golang/1.16.2
#xsmall
for i in {1..5}
do
    go run benchmark.go s xsmall >> time.txt
done

for j in {1..6}
do
    for i in {1..5}
    do
        go run benchmark.go p xsmall $(($j * 2)) >> time.txt
    done
done

# #small
# for i in {1..5}
# do
#     go run benchmark.go s small >> time.txt
# done

# for j in {1..6}
# do
#     for i in {1..5}
#     do
#         go run benchmark.go p small $(($j * 2)) >> time.txt
#     done
# done

# #medium
# for i in {1..5}
# do
#     go run benchmark.go s medium >> time.txt
# done

# for j in {1..6}
# do
#     for i in {1..5}
#     do
#         go run benchmark.go p medium $(($j * 2)) >> time.txt
#     done
# done

# #large
# for i in {1..5}
# do
#     go run benchmark.go s large >> time.txt
# done

# for j in {1..6}
# do
#     for i in {1..5}
#     do
#         go run benchmark.go p large $(($j * 2)) >> time.txt
#     done
# done

# #xlarge
# for i in {1..5}
# do
#     go run benchmark.go s xlarge >> time.txt
# done

# for j in {1..6}
# do
#     for i in {1..5}
#     do
#         go run benchmark.go p xlarge $(($j * 2)) >> time.txt
#     done
# done

# module load python
python3 generateGraphs.py