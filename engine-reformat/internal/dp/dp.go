package dp

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	cmn "pipeline/internal/common"
	"strconv"

	//"strings"
	"time"
)

const channelSize = 5000

type in_comm struct {
	// read-only channels
	Edge          <-chan cmn.Edge // Edges channel
	Front_Channel <-chan in_comm  // in_comm channel
}

func Sink(in_alerts <-chan cmn.Alert, in_log <-chan cmn.Edge) {
	fmt.Println("SINK PROCESS - LAUNCHED")
	// TOCHECK: Create results output files: one for each kind of fraud pattern (?)
	// TODO: For the moment only 1 kind of pattern
	file_fp_1, err := os.Create("../output/outPattern1.txt")
	cmn.CheckError(err)
	defer file_fp_1.Close()

	// Logging file
	file_log, err := os.Create("../output/log.txt")
	cmn.CheckError(err)
	defer file_log.Close()

	for {
		select {
		case alert, ok := <-in_alerts:
			if ok {
				fmt.Println("Sink - alert!: ", alert)
				cmn.PrintAlertVerbose(alert)
				cmn.PrintAlertOnFile(alert, file_fp_1)
			}
		case event, ok := <-in_log:
			if ok {
				cmn.PrintEventOnFile(event, file_log)
			}
		}
	}
}

func Generator(in_edges <-chan cmn.Edge, out_alerts chan<- cmn.Alert, out_log chan<- cmn.Edge) {
	fmt.Println("G - creation")
	for {
		select {
		case edge := <-in_edges:
			//cmn.PrintEdge("G - edge arrived: ", edge)
			// spawn a filter
			new_edge_ch := make(chan cmn.Edge, channelSize)
			go filter(edge, in_edges, new_edge_ch, out_alerts)
			// set the new input channels of the generator
			in_edges = new_edge_ch
		default:
			// TODO
			fmt.Println("G: Default case")
		}
	}
	// TODO: Close channels
}

func filter(edge cmn.Edge, in_edge <-chan cmn.Edge, out_edge chan<- cmn.Edge, out_alerts chan<- cmn.Alert) {

	// filter id: is the Card unique identifier
	var id string = edge.Number_id

	fmt.Println("F ", id, "- creation")

	int_edge := make(chan cmn.Edge, channelSize)
	int_time := make(chan time.Time) // synchronous
	// TOCHECK: Avoid this channel being blocking (?) Does it make sense?
	int_stop := make(chan bool) // synchronous

	go filter_worker(edge, int_edge, int_time, int_stop, out_alerts)

	for {
		select {
		case edge := <-in_edge:
			cmn.PrintEdge("F - edge arrived: ", edge)
			if edge.Number_id == id {
				int_edge <- edge
			} else {
				out_edge <- edge
			}
		default:
			// TODO
			fmt.Println("F: Default case")
		}
		// TODO: Close channels
	}
}

func filter_worker(initial_edge cmn.Edge, int_edge <-chan cmn.Edge, int_time <-chan time.Time, int_stop chan<- bool, out_alerts chan<- cmn.Alert) {

	cmn.PrintEdge("FW creation - edge arrived: ", initial_edge)
	var subgraph *cmn.Graph = cmn.NewGraph() // Explicit declaration
	isStart := initial_edge.IsStart()
	if !isStart {
		log.Fatalf("Error: AddEdge ->  Initial edge of the filter is not of type tx-start")
	}
	subgraph.AddEdge(initial_edge)
	subgraph.PrintIds()

	// TODO: this goroutine dies alone after its father (the filter) dies?
	// -> it is the only process with which it has communication / is connected
	for {
		select {
		case new_edge := <-int_edge:
			cmn.PrintEdge("FW - edge arrived: ", new_edge)
			// NOTE: New -> with 2 edges per tx
			// 1. Identify if it is start or end edge
			isStart := new_edge.IsStart()
			if isStart {
				// start edge
				// 1. Check fraud
				isFraud, alert := subgraph.CheckFraud(new_edge)
				if isFraud {
					out_alerts <- alert
				}
				fmt.Println("........................................")
				// 2. Add to the subgraph
				subgraph.AddEdge(new_edge)
			} else {
				fmt.Println("Is end")
				subgraph.CompleteEdge(new_edge)
			}
			subgraph.Print()
		}

	}
}

func Source(istream string, out_edges chan<- cmn.Edge) {

	// input stream file
	file, err := os.Open(istream)
	cmn.CheckError(err)
	defer file.Close()
	cmn.CheckError(err)

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read() // Read and discard the header line
	cmn.CheckError(err)

	for {

		tx, err := reader.Read()
		if err == io.EOF {
			fmt.Println("End of stream...")
			// TODO: Create op = EOF
			break
		}
		cmn.CheckError(err)

		// conversions
		tx_id_64, err := strconv.ParseInt(tx[0], 10, 32) // 10: base (decimal) & 32: bit-size (int32)
		cmn.CheckError(err)
		// still the type returned is int64 -> convert to int32
		tx_id := int32(tx_id_64)

		tx_start, err := time.Parse(cmn.Time_layout, tx[3])
		cmn.CheckError(err)

		// Check if tx_end field is empty
		// From: https://pkg.go.dev/time#Time
		// The zero value of type Time is January 1, year 1, 00:00:00.000000000 UTC. As this time
		// is unlikely to come up in practice, the Time.IsZero method gives a simple way of detecting
		// a time that has not been initialized explicitly.
		var tx_end time.Time
		if tx[4] != "" {
			tx_end, err = time.Parse(cmn.Time_layout, tx[4])
			cmn.CheckError(err)
		}

		var tx_amount_32 float32
		if tx[5] != "" {
			tx_amount, err := strconv.ParseFloat(tx[5], 32)
			cmn.CheckError(err)
			tx_amount_32 = float32(tx_amount)
		}

		edge := cmn.Edge{
			Number_id: tx[1],
			ATM_id:    tx[2],
			Tx_id:     tx_id,
			Tx_start:  tx_start,
			Tx_end:    tx_end,
			Tx_amount: tx_amount_32,
		}

		cmn.PrintEdgeComplete("", edge)
		out_edges <- edge

		// TODO-REMOVE: -- Only for testing/debugging purposes --
		//  Sleep time for debugging to slow down the flux of transactions
		// Leave without this sleep / change it
		time.Sleep(500 * time.Millisecond)

	}

}
