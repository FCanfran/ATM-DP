/*
Entry point of the program
*/

package main

import (
	"pipeline/internal/connection"
)

func main() {

	// Connection to static gdb
	connection.SafeConnect()

	/*
		// obtain stream fileName from args
		istream := os.Args[1]

		fmt.Println("Reading stream from: ", istream)

		// start the pipeline and give the stream edge by edge
		dp.Start(istream)
	*/
}
