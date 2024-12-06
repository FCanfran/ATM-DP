#!/bin/bash -l
#
#SBATCH -J exp-1c-16f
#SBATCH -o exp-1c-16f."%j".out
#SBATCH -e exp-1c-16f."%j".err
#
#SBATCH --mail-user fernando.martin.canfran@estudiantat.upc.edu
#SBATCH --mail-type=ALL
#
#SBATCH --mem=1024M
#SBATCH -c 1
#SBATCH -p short

../cmd/main exps-descriptions/small/1c/1c-16f.csv
