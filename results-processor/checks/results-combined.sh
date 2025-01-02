#!/bin/bash

if [ "$#" -ne 4 ]; then
  echo "Usage: $0 <resultsDirectoryPath> <TEST(name)> <DO_JOIN(0:no,1:yes)> <num_interactions>"
  exit 1
fi

# Directory passed as a parameter
RESULTS_DIR=$1
TEST_NAME=$2
DO_JOIN=$3
NUM_INTERACTIONS=$4

rm -r TOPLOT
mkdir TOPLOT
cp -r "$RESULTS_DIR/"*"-avg" TOPLOT

python3 dieffpy-combined.py TOPLOT $TEST_NAME $DO_JOIN $NUM_INTERACTIONS > dieffpy-out.txt
