#!/bin/bash -l
#
#SBATCH -J exp-16c
#SBATCH -o exp-16c."%j".out
#SBATCH -e exp-16c."%j".err
#
#SBATCH --mail-user fernando.martin.canfran@estudiantat.upc.edu
#SBATCH --mail-type=ALL
#
#SBATCH --mem=65536M
#SBATCH -c 8
#SBATCH -p short

./results-fix-filters-main.sh results-medium-7-0.03-edit 7-0.03