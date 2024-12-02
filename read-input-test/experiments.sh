#!/bin/bash

if [ $# -lt 3 ]; then
  echo "Usage: $0 <numApproach{1|2|3}> <stream_file> <max power of 10>"
  exit 1
fi

num_approach=$1
input_file=$2
max_power=$3

# outputfilename
base_name=$(basename "$input_file" .csv)

if [ "$num_approach" -eq 1 ]; then

  output_file="${base_name}-1-apache-arrow.csv"

  if [ -f "$output_file" ]; then
    echo "Deleting existing output file: $output_file"
    rm "$output_file"
  fi

  # Loop from 0 to max_power
  for ((power=0; power<=max_power; power++)); do

    chunk_size=$((10 ** power)) # Calculate 10^power
    echo "Running for chunk size: $chunk_size"

    for ((i=1; i<=100; i++)); do
      echo "Experiment #$i for chunk size: $chunk_size"
      go run cmd/1-apache-arrow/main.go "$input_file" "$chunk_size" "$output_file"
    done

  done

elif [ "$num_approach" -eq 2 ]; then

  output_file="${base_name}-2-apache-arrow.csv"

  if [ -f "$output_file" ]; then
    echo "Deleting existing output file: $output_file"
    rm "$output_file"
  fi

  # Loop from 0 to max_power
  for ((power=0; power<=max_power; power++)); do

    chunk_size=$((10 ** power)) # Calculate 10^power
    echo "Running for chunk size: $chunk_size"

    for ((i=1; i<=100; i++)); do
      echo "Experiment #$i for chunk size: $chunk_size"
      go run cmd/2-apache-arrow/main.go "$input_file" "$chunk_size" "$output_file"
    done

  done

elif [ "$num_approach" -eq 3 ]; then


  output_file="${base_name}-3-apache-arrow.csv"

  if [ -f "$output_file" ]; then
    echo "Deleting existing output file: $output_file"
    rm "$output_file"
  fi

  # Loop from 0 to max_power
  for ((power=0; power<=max_power; power++)); do

    chunk_size=$((10 ** power)) # Calculate 10^power
    echo "Running for chunk size: $chunk_size"
        
    for ((i=1; i<=100; i++)); do
      echo "Experiment #$i for chunk size: $chunk_size"
      go run cmd/3-apache-arrow/main.go "$input_file" "$chunk_size" "$output_file"
    done

  done

else
  echo "Invalid numApproach: $num_approach. Must be 1, 2, or 3."
  exit 1
fi


