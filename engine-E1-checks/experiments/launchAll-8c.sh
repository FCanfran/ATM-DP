#!/bin/bash -l
#
#SBATCH -J exp-8c
#SBATCH -o exp-8c."%j".out
#SBATCH -e exp-8c."%j".err
#
#SBATCH --mail-user fernando.martin.canfran@estudiantat.upc.edu
#SBATCH --mail-type=ALL
#
#SBATCH --mem=16384M
#SBATCH -c 8
#SBATCH -p short


if [ $# -ne 2 ]; then
    echo "Usage: $0 <exps-descriptions-directory> <experiment-execTimes>"
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
#echo "compilation..."
#go build -o ../cmd/main ../cmd/main.go

# 2. run all the experiments, one after the other
for csv_description_file in $(ls "$directory"/*.csv | sort -V); do # sort -V to respect numerical order
    if [ -f "$csv_description_file" ]; then
        filename=$(basename "$csv_description_file") 
        base="${filename%.csv}"        
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
            echo "Executing experiment $base run $i..."  
            echo "___________________________________________________________________________________________________________"
            echo 
            ../cmd/main "$csv_description_file" # exec
            rm -r "$outdir-$i"
            mv $outdir "$outdir-$i" # rename - appending the label of the corresponding run
            
            # calculate the Mean Response Time - based on the trace.csv and add it to the metrics.csv
            python3 calculate_mrt.py "$outdir-$i/trace.csv" "$outdir-$i/metrics.csv"

            # append the csv metrics and traces files into the avg files
            # metrics
            if [ -s "$outdir-$i/metrics.csv" ]; then  # file exists and is non-empty?
                if [ ! -f "$metrics_outfile" ]; then
                    head -n 1 "$outdir-$i/metrics.csv" > "$metrics_outfile" # add header if does not exist
                fi
                tail -n +2 "$outdir-$i/metrics.csv" >> "$metrics_outfile" # append, excluding the header
            fi
            
            # traces
            if [ -s "$outdir-$i/trace.csv" ]; then  # file exists and is non-empty?
                if [ ! -f "$trace_outfile" ]; then
                    head -n 1 "$outdir-$i/trace.csv" > "$trace_outfile" # add header if does not exist
                fi
                tail -n +2 "$outdir-$i/trace.csv" >> "$trace_outfile" # append, excluding the header
            fi

            if [ $i -ne 1 ]; then
                # Optional: remove the current directory - we will keep only the -avg one
                echo "rm -r $outdir-$i"
                rm -r "$outdir-$i"
            fi
        
        done

        # average metrics & trace csvs
        python3 average_metrics.py $metrics_outfile
        python3 average_traces.py $trace_outfile

        
    else
        echo "No .sh files found in $directory."
    fi
done
