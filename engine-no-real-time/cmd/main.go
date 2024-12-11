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
	maxProcsBefore := runtime.GOMAXPROCS(0)
	maxProcsNow := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	fmt.Println("maxProcsBefore: ", maxProcsBefore, "maxProcsNow: ", maxProcsNow, " numCPU: ", numCPU)

	/*
		// create go execution traces output files
		var cpuprofile string = cmn.OutDirName + "/cpu.pprof"
		var memprofile string = cmn.OutDirName + "/mem.pprof"
		var tracefile string = cmn.OutDirName + "/trace-go.out"
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		f, err = os.Create(tracefile)
		if err != nil {
			log.Fatal(err)
		}
		if err := trace.Start(f); err != nil {
			log.Fatalf("failed to start trace: %v", err)
		}
		defer trace.Stop()
	*/

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
	go dp.Stream(cmn.StreamFileName, stream_ch)
	// launch Source, Generator and Sink goroutines
	go dp.Source(stream_ch, event_ch)
	go dp.Generator(event_ch, alert_ch, out_event_ch)
	go dp.Sink(start, alert_ch, out_event_ch, endchan)

	<-endchan
	t := time.Since(start)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")
	fmt.Println("Finish Program")

	// finish connection to static gdb
	connection.CloseConnection(ctx)

	/*
		f, err = os.Create(memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	*/

}
