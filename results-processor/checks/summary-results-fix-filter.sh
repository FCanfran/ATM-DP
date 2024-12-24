#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Usage: $0 <directory> <TEST>"
    exit 1
fi

directory="$1"
TEST="$2"

# at this point -> "output" directory with all the results -> labeled as -avg directories
# - produce plots and diefficiency results with the diefpy library program
# - move the result subdirectories of the "output" directory

# move all the -avg directories into the same directory and run the dieffpy program there
outdirallavg="$directory/avg-all"
rm -r $outdirallavg
mkdir -p $outdirallavg

# Find directories ending with -avg and copy them to the target directory
find $directory -type d -name "*-avg" | while read -r dir; do
    echo "$dir"
    cp -r "$dir" "$outdirallavg/"
    echo "Copied $dir to $outdirallavg/"
done


outfiledieff="$outdirallavg/dieffpy-out.txt"
python3 dieffpy-filters.py $outdirallavg $TEST 1 > $outfiledieff


# create an output dir called as TEST
rm -r $TEST
mkdir $TEST
# put the results in this directory
mv "$outdirallavg/plots" $TEST
mv "$outdirallavg/dieffpy-out.txt" $TEST
mv "$outdirallavg/metrics.csv" $TEST
mv "$outdirallavg/trace.csv" $TEST