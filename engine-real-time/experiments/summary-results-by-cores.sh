#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Usage: $0 <directory> <TEST>"
    exit 1
fi

directory="$1"
TEST="$2"

# for each of the output/*-avg directories take the metrics.csv & trace.csv files and apply averaging
for metrics_outfile in $(ls "$directory"/*-avg/metrics.csv | sort -V); do # sort -V to respect numerical order
    echo $metrics_outfile
    python3 average_metrics.py $metrics_outfile
done

for trace_outfile in $(ls "$directory"/*-avg/trace.csv | sort -V); do # sort -V to respect numerical order
    echo $trace_outfile
    python3 average_traces.py $trace_outfile
done

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
python3 dieffpy-cores.py $outdirallavg $TEST 1 > $outfiledieff

#outdirTest="out-$TEST"
#rm -r $outdirTest
#mv "./output" $outdirTest