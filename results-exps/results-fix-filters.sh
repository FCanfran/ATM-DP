#!/bin/bash

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <results-directory> <testName>"
  exit 1
fi

# Directory passed as a parameter
RESULTS_DIR=$1
TEST_NAME=$2

filter_values=("1f" "2f" "5f" "10f" "20f" "40f" "100f" "200f" "500f" "1000f" "2000f")

for x in "${filter_values[@]}"; do
  echo "Getting plots and results for $x..."
  rm -r TOPLOT/*
  cp -r "$RESULTS_DIR/"*"-$x-"* TOPLOT
  ./summary-results-fix-filter.sh TOPLOT "$TEST_NAME"
  mv "$TEST_NAME" "$x"
  echo "Done for $x"
done
