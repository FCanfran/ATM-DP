#!/bin/bash

if [ $# -ne 3 ]; then
    echo "Usage: $0 <exps-directory> <experiment-execTimes> <TEST>"
    exit 1
fi

# Assign the directory from the command line argument
directory="$1"
execTimes="$2"
TEST="$3"

# Check if the provided directory exists
if [ ! -d "$directory" ]; then
    echo "Directory $directory does not exist."
    exit 1
fi


# 1. compilation
echo "compilation..."
go build -o ../cmd/main ../cmd/main.go

# 2. launch all the experiments scripts, one after the other (in the exps-scrips directory)
for script in $(ls "$directory"/*.sh | sort -V); do # sort -V to respect numerical order
    if [ -f "$script" ]; then
        filename=$(basename "$script") 
        base="${filename%.sh}"        
        outdir="output/$base"
        outdiravg="output/$base-avg"

        # create output directory for the averaged results
        rm -r $outdiravg 
        mkdir -p "$outdiravg"
        echo "Directory '$outdiravg' created."
        
        # output averaged files: metrics.csv and trace.csv
        metrics_outfile="$outdiravg/metrics.csv"
        trace_outfile="$outdiravg/trace.csv"
        rm -f "$metrics_outfile"
        rm -f "$trace_outfile"

        # execute each experiment execTimes
        for ((i = 1; i <= execTimes; i++)); do
            echo "___________________________________________________________________________________________________________"
            echo 
            echo "Executing $script run $i..."  
            echo "___________________________________________________________________________________________________________"
            echo 
            bash "$script" # outdir is the script output directory
            rm -r "$outdir-$i"
            mv $outdir "$outdir-$i" # rename - appending the label of the corresponding run
            # append the csv metrics and traces files into the avg files
            # metrics
            if [ ! -f "$metrics_outfile" ]; then
                head -n 1 "$outdir-$i/metrics.csv" > "$metrics_outfile" # add header if does not exist
            fi
            tail -n +2 "$outdir-$i/metrics.csv" >> "$metrics_outfile" # append, excluding the header

            # traces
            if [ ! -f "$trace_outfile" ]; then
                head -n 1 "$outdir-$i/trace.csv" > "$trace_outfile" # add header if does not exist
            fi
            tail -n +2 "$outdir-$i/trace.csv" >> "$trace_outfile" # append, excluding the header

        done
        # average the results of this experiment in a single output - averaged - directory
        # averaged metrics.csv
        # averaged trace.csv
        python3.10 average_metrics.py $metrics_outfile
        python3.10 average_traces.py $trace_outfile
    else
        echo "No .sh files found in $directory."
    fi
done

# at this point -> "output" directory with all the results -> labeled as -avg directories
# - produce plots and diefficiency results with the diefpy library program
# - move the result subdirectories of the "output" directory

# move all the -avg directories into the same directory and run the dieffpy program there
outdirallavg="output/avg-all"
rm -r $outdirallavg
mkdir -p $outdirallavg

# Find directories ending with -avg and move them to the target directory
find . -type d -name "*-avg" | while read -r dir; do
    echo "$dir"
    mv "$dir" "$outdirallavg/"
    echo "Moved $dir to $outdirallavg/"
done


outfiledieff="$outdirallavg/dieffpy-out.txt"
python3.10 dieffpy.py $outdirallavg $TEST > $outfiledieff

outdirTest="out-$TEST"
rm -r $outdirTest
mv "./output" $outdirTest