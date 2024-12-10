package main

import (
	dp "dp_mst/internal/dp"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

func main() {
	//-----------------------
	//
	var istream string
	//Pprof variables -----------------------
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
	var tracefile = flag.String("tracefile", "", "write trace execution to `file`")
	flag.StringVar(&istream, "file", "", "Specify input file. Default is emptyfile")
	flag.Parse()

	//---------------------------------------
	// Filter size (number of "root nodes" stored by each filter)
	var Fsize int
	fmt.Print("Number of nodes per filter: ")
	fmt.Scan(&Fsize)

	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	fmt.Println("maxProcs: ", maxProcs, " numCPU: ", numCPU)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *tracefile != "" {
		f, err := os.Create(*tracefile)
		if err != nil {
			log.Fatal(err)
		}
		if err := trace.Start(f); err != nil {
			log.Fatalf("failed to start trace: %v", err)
		}
		defer trace.Stop()
	}

	dp.Start(istream, Fsize)

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
