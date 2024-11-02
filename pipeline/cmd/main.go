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
	// start connection to static gdb
	connection.SafeConnect()

	// obtain stream fileName from args
	istream := os.Args[1]

	// start the pipeline and give the stream edge by edge
	// TODO: Llamar SOURCE en vez de Start
	dp.Start(istream)

}
