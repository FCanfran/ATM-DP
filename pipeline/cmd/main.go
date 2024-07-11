/*
Entry point of the program
*/

package main

import (
	"pipeline/internal/dp"
	"fmt"
	"os"
)

func main() {

	// obtain stream fileName from args
	istream := os.Args[1]

	fmt.Println("Reading stream from: ", istream)

	// start the pipeline and give the stream edge by edge
	dp.Start(istream)
}
