#!/bin/bash

for x in $(seq 10000 10000 1000000); do
    go run csvGenerator.go $x
done
