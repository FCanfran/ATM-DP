#!/bin/bash

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <results-directory> <testName>"
  exit 1
fi

# Directory passed as a parameter
RESULTS_DIR=$1
TEST_NAME=$2
core_values=("1c" "2c" "4c" "8c" "16c")

rm -r TOPLOT
mkdir TOPLOT

for x in "${core_values[@]}"; do
  echo "Getting plots and results for $x..."
  rm -r TOPLOT/*
  cp -r "$RESULTS_DIR/$x-"*"-avg" TOPLOT
  ./results-fix-cores-a.sh TOPLOT $TEST_NAME-$x
  mv $TEST_NAME-$x $x
  echo "Done for $x"
done
