/*
Entry point of the program
*/

package main

import (
	"fmt"
	"os"
	cmn "pipeline/internal/common"
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

	// creation of needed channels
	// channel to pass from the read input to the pipeline
	edges_ch := make(chan cmn.Edge, cmn.ChannelSize)
	// alerts channel
	alerts_ch := make(chan cmn.Alert, cmn.ChannelSize)
	// Log channel: to register all the events generated in the engine. Bidirectional.
	// Registering in the sink
	// TOCHECK: For the moment only edges through it
	log_ch := make(chan cmn.Edge, cmn.ChannelSize)
	// Ending channel
	endchan := make(chan bool, 1) //channel transporting sorted lists (graph/kmst)

	// launch Source, Generator and Sink goroutines
	go dp.Source(istream, edges_ch)
	go dp.Generator(edges_ch, alerts_ch, log_ch)
	go dp.Sink(alerts_ch, log_ch)

	// TODO: Crear channel para esperar la terminación
	<-endchan
}
