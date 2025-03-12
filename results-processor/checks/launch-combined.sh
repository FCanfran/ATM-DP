#!/bin/bash -l
#
#SBATCH -J exp-16c
#SBATCH -o exp-16c."%j".out
#SBATCH -e exp-16c."%j".err
#
#SBATCH --mail-user fernando.martin.canfran@estudiantat.upc.edu
#SBATCH --mail-type=ALL
#
#SBATCH --mem=32768M 
#SBATCH -c 8
#SBATCH -p short

./results-combined.sh results-small-120-0.02 120-0.02 1 321500

