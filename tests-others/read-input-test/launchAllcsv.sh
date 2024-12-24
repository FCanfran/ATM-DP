#!/bin/bash


# compilation
# ------------------------------------------------------------------------- # 
cd cmd/3-csv-encoding || exit 1
go build main.go
# return to the original directory
cd - || exit 1

cd cmd/4-csv-encoding || exit 1
go build main.go
# return to the original directory
cd - || exit 1
# ------------------------------------------------------------------------- # 

./experiments-csv-encoding.sh streams/10-4.csv 
./experiments-csv-encoding.sh streams/10-5.csv 
./experiments-csv-encoding.sh streams/10-6.csv 
./experiments-csv-encoding.sh streams/10-7.csv 