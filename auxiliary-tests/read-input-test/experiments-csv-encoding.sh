#!/bin/bash

if [ $# -lt 1 ]; then
  echo "Usage: $0 <stream_file>"
  exit 1
fi

input_file=$1
# outputfilename
base_name=$(basename "$input_file" .csv)

# ------------------------------------------------------------------------- # 
# 3-csv/encoding - worker
output_file="3-csv-encoding.csv"
chunk_size=100
for ((i=1; i<=20; i++)); do
      echo "Experiment #$i 3-csv/encoding -- worker"
      cmd/3-csv-encoding/main "$input_file" "$chunk_size" "$output_file"
done

# ------------------------------------------------------------------------- # 
# 4-csv/encoding - no worker
output_file="4-csv-encoding.csv"
# NOTE: no chunk size in reality / no chunks! ---- 1 by 1
for ((i=1; i<=20; i++)); do
      echo "Experiment #$i 4-csv/encoding -- no worker"
      cmd/4-csv-encoding/main "$input_file" "$chunk_size" "$output_file"
done
