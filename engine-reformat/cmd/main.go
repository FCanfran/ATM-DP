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
	// dedicated Edge channel to pass tx read from input to the pipeline
	edge_ch := make(chan cmn.Edge, cmn.ChannelSize)
	// event channel
	event_ch := make(chan cmn.Event, cmn.ChannelSize)
	// alerts channel
	alert_ch := make(chan cmn.Alert, cmn.ChannelSize)
	// out_event_ch channel: direct event channel between Generator and Sink.
	// --> all kinds of events except the alerts
	out_event_ch := make(chan cmn.Event, cmn.ChannelSize)
	// Ending channel
	endchan := make(chan struct{})

	// launch Source, Generator and Sink goroutines
	go dp.Source(istream, edge_ch, event_ch)
	go dp.Generator(edge_ch, event_ch, alert_ch, out_event_ch)
	go dp.Sink(alert_ch, out_event_ch, endchan)

	<-endchan
	fmt.Println("Finish Program")
}
