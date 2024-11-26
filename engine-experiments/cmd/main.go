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
	"time"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <transactionFileName> <scalingFactor>")
		return
	}

	// obtain stream fileName from args
	istream := os.Args[1]
	// scaling factor (T_new / T_original)
	// - of the time interval of the input transaction stream: [0,1]
	cmn.SetScaleFactor(os.Args[2])
	// TODO: Define exp rootname after the input stream filename
	cmn.SetRootFileName("exp-1")

	// start connection to static gdb
	ctx := connection.SafeConnect()

	// creation of needed channels
	// real-time input stream channel
	stream_ch := make(chan cmn.Event, cmn.ChannelSize)
	// event channel
	event_ch := make(chan cmn.Event, cmn.ChannelSize)
	// alerts channel
	alert_ch := make(chan cmn.Alert, cmn.ChannelSize)
	// out_event_ch channel: direct event channel between Generator and Sink.
	// --> all kinds of events except the alerts
	out_event_ch := make(chan cmn.Event, cmn.ChannelSize)
	// Ending channel
	endchan := make(chan struct{})

	start := time.Now()
	// launch Stream goroutines - to provide the input in real-time
	go dp.Stream(istream, stream_ch)
	// launch Source, Generator and Sink goroutines
	go dp.Source(stream_ch, event_ch)
	go dp.Generator(event_ch, alert_ch, out_event_ch)
	go dp.Sink(start, alert_ch, out_event_ch, endchan)

	<-endchan
	t := time.Since(start)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")
	fmt.Println("Finish Program")

	// TODO: finish connection to static gdb
	connection.CloseConnection(ctx)

}
