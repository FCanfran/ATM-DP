/*
Entry point of the program
*/

package main

import (
	"fmt"
	"os"
	"pipeline/internal/connection"
	"pipeline/internal/dp"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <transactionFileName>")
		return
	}
	// Connection to static gdb
	connection.SafeConnect()

	// obtain stream fileName from args
	istream := os.Args[1]

	fmt.Println("Reading stream from: ", istream)

	// start the pipeline and give the stream edge by edge
	dp.Start(istream)

}
