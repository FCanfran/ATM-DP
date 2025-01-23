package main

import (
	"fmt"
	"os"
	cmn "pipeline/internal/common"
	"pipeline/internal/connection"
	"pipeline/internal/dp"
	"runtime"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <executionDescriptionFile>")
		return
	}

	cmn.ReadExecDescriptionFile(os.Args[1])

	// start connection to static gdb
	ctx := connection.SafeConnect()

	// golang max processors settings
	// GOMAXPROCS sets the maximum number of CPUs that can be executing simultaneously and returns the previous setting.
	// It defaults to the value of runtime.NumCPU. If n < 1, it does not change the current setting.
	maxProcsBefore := runtime.GOMAXPROCS(0)
	maxProcsNow := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	fmt.Println("maxProcsBefore: ", maxProcsBefore, "maxProcsNow: ", maxProcsNow, " numCPU: ", numCPU)

	// creation of needed channels
	// real-time input stream channel
	stream_ch := make(chan cmn.Event, cmn.ChannelSize)
	// event channel
	event_ch := make(chan cmn.Event, cmn.ChannelSize)
	// alerts channel
	check_ch := make(chan cmn.CheckResult, cmn.ChannelSize)
	// out_event_ch channel: direct event channel between Generator and Sink.
	// --> all kinds of events except the alerts
	out_event_ch := make(chan cmn.Event, cmn.ChannelSize)
	// Ending channel
	endchan := make(chan struct{})

	start := time.Now()
	// launch Stream goroutines - to provide the input in real-time
	go dp.Stream(cmn.StreamFileName, stream_ch)
	// launch Source, Generator and Sink goroutines
	go dp.Source(start, stream_ch, event_ch)
	go dp.Generator(event_ch, check_ch, out_event_ch)
	go dp.Sink(start, check_ch, out_event_ch, endchan)

	<-endchan
	t := time.Since(start)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")
	fmt.Println("Finish Program")

	// finish connection to static gdb
	connection.CloseConnection(ctx)

}
