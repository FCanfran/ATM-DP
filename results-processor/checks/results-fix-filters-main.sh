#!/bin/bash

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <results-directory> <testName>"
  exit 1
fi

# Directory passed as a parameter
RESULTS_DIR=$1
TEST_NAME=$2
#filter_values=("1f" "2f" "5f" "10f" "20f" "40f" "100f" "200f" "500f" "1000f" "2000f")
filter_values=("5f" "10f" "100f" "250f" "500f" "1000f" "2000f" "5000f" "10000f")


rm -r TOPLOT-FILTERS
mkdir TOPLOT-FILTERS

for x in "${filter_values[@]}"; do
  echo "Getting plots and results for $x..."
  rm -r TOPLOT-FILTERS/*
  cp -r "$RESULTS_DIR/"*"-$x-avg" TOPLOT-FILTERS
  ./results-fix-filters-a.sh TOPLOT-FILTERS $TEST_NAME-$x
  mv $TEST_NAME-$x $x
  echo "Done for $x"
done
