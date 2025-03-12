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

rm -r "TOPLOT-COMBINED-$TEST_NAME"
mkdir "TOPLOT-COMBINED-$TEST_NAME"
cp -r "$RESULTS_DIR/"*"-avg" "TOPLOT-COMBINED-$TEST_NAME"

python3 dieffpy-combined.py "TOPLOT-COMBINED-$TEST_NAME" $TEST_NAME $DO_JOIN $NUM_INTERACTIONS > "TOPLOT-COMBINED-$TEST_NAME/dieffpy-out.txt"
