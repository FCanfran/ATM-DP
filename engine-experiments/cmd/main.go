/*
Entry point of the program
*/

package main

import (
	"fmt"
	"os"
	"pipeline/internal/dp"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <transactionFileName> <scalingFactor>")
		return
	}

	istream := os.Args[1]
	endchan := make(chan struct{})
	go dp.Stream(istream)
	<-endchan

	/*
		// start connection to static gdb
		ctx := connection.SafeConnect()

		// obtain stream fileName from args
		istream := os.Args[1]
		// scaling factor (T_new / T_original) - of the time interval of the input transaction stream: [0,1]
		cmn.SetScaleFactor(os.Args[2])

		// creation of needed channels
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
		// launch Source, Generator and Sink goroutines
		go dp.Source(istream, event_ch)
		go dp.Generator(event_ch, alert_ch, out_event_ch)
		go dp.Sink(start, alert_ch, out_event_ch, endchan)

		<-endchan
		t := time.Since(start)
		fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")
		fmt.Println("Finish Program")

		// TODO: finish connection to static gdb
		connection.CloseConnection(ctx)
	*/

}
