#!/bin/bash -l
#
#SBATCH -J exp-1
#SBATCH -o exp-1."%j".out
#SBATCH -e exp-1."%j".err
#
#SBATCH --mail-user fernando.martin.canfran@estudiantat.upc.edu
#SBATCH --mail-type=ALL
#
#SBATCH --mem=1024M
#SBATCH -c 1
#SBATCH -p short

# firstly we will compile the .c program
#gcc -o helloworld.exe helloworld.c

# Secondly we will run the program
#go run cmd/main.go cmd/upt-all.csv > out.txt
#cmd/main cmd/upt-all.csv > out.txt
go run cmd/main.go exps/exp-small.csv > out.txt
