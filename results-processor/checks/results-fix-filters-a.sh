#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Usage: $0 <directory> <TEST>"
    exit 1
fi

directory="$1"
TEST="$2"

outfiledieff="$directory/dieffpy-out.txt"
python3 dieffpy-filters.py $directory $TEST 1 > $outfiledieff

# create an output dir called as TEST
rm -r $TEST
mkdir $TEST
# put the results in this directory
mv "$directory/plots" $TEST
mv "$directory/dieffpy-out.txt" $TEST
mv "$directory/metrics.csv" $TEST
mv "$directory/trace.csv" $TEST