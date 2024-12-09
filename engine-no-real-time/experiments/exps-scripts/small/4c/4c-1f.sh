#!/bin/bash -l
#
#SBATCH -J exp-4c-1f
#SBATCH -o exp-4c-1f."%j".out
#SBATCH -e exp-4c-1f."%j".err
#
#SBATCH --mail-user fernando.martin.canfran@estudiantat.upc.edu
#SBATCH --mail-type=ALL
#
#SBATCH --mem=1024M
#SBATCH -c 4
#SBATCH -p short

../cmd/main exps-descriptions/small/4c/4c-1f.csv
