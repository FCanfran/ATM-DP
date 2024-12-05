#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Usage: $0 <exps-directory> <experiment-execTimes>"
    exit 1
fi

# Assign the directory from the command line argument
directory="$1"
execTimes="$2"

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
        # execute each experiment execTimes
        for ((i = 1; i <= execTimes; i++)); do
            echo "Executing $script run $i..."  
            bash "$script"
            mv $outdir "$outdir-$i" # rename - appending the label of the corresponding run
        done
        # average the results of this experiment in a single output - averaged - directory
    else
        echo "No .sh files found in $directory."
    fi
    exit
done

# at this point -> "output" directory with all the results 
# - produce plots and diefficiency results with the diefpy library program
# - move the result subdirectories of the "output" directory

