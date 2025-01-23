#!/bin/bash


cd cmd/1-apache-arrow || exit 1
go build main.go
# return to the original directory
cd - || exit 1

cd cmd/2-apache-arrow || exit 1
go build main.go
# return to the original directory
cd - || exit 1

cd cmd/3-csv-encoding || exit 1
go build main.go
# return to the original directory
cd - || exit 1

./experiments.sh 1 streams/10-4.csv 4
./experiments.sh 1 streams/10-5.csv 5
./experiments.sh 1 streams/10-6.csv 6

./experiments.sh 2 streams/10-4.csv 4
./experiments.sh 2 streams/10-5.csv 5
./experiments.sh 2 streams/10-6.csv 6

./experiments.sh 3 streams/10-4.csv 4
./experiments.sh 3 streams/10-5.csv 5
./experiments.sh 3 streams/10-6.csv 6