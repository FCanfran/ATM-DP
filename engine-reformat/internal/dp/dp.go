package dp

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	cmn "pipeline/internal/common"
	"strconv"

	//"strings"
	"time"
)

const channelSize = 5000

func Sink(in_alert <-chan cmn.Alert, in_event <-chan cmn.Event, endchan chan<- struct{}) {
	fmt.Println("Sink - creation")
	// TOCHECK: Create results output files: one for each kind of fraud pattern (?)
	// TODO: For the moment only 1 kind of pattern
	file_fp_1, err := os.Create("../output/outPattern1.txt")
	cmn.CheckError(err)
	defer file_fp_1.Close()

	// Logging file
	file_log, err := os.Create("../output/log.txt")
	cmn.CheckError(err)
	defer file_log.Close()

Loop:
	for {
		select {
		case alert, ok := <-in_alert:
			if ok {
				fmt.Println("Sink - alert!: ", alert)
				cmn.PrintAlertVerbose(alert)
				cmn.PrintAlertOnFile(alert, file_fp_1)
			}
		case event, ok := <-in_event:
			if ok {
				// TODO: Print the event and not (only) the edge associated?
				cmn.PrintEventOnFile(event.E, file_log)
			}
			switch event.Type {
			case cmn.EOF:
				fmt.Println("Sink - EOF event")
				// finish the Sink
				break Loop

			}
		}
	}
	endchan <- struct{}{}
}

func Generator(
	in_edge <-chan cmn.Edge,
	in_event <-chan cmn.Event,
	out_alert chan<- cmn.Alert,
	out_event chan<- cmn.Event) {

	fmt.Println("G - creation")
Loop:
	for {
		// NOTE: "select" so to leave it prepared when re-introducing the reconnection channels
		// --> we want to be able to read from multiple input channels at the same time
		select {
		case edge, ok := <-in_edge:
			if !ok {
				// TODO: Manage the error properly
				fmt.Println("G - !ok in in_edge channel")
			}
			cmn.PrintEdge("G - edge arrived: ", edge)
			// spawn a filter
			new_edge_ch := make(chan cmn.Edge, cmn.ChannelSize)
			new_event_ch := make(chan cmn.Event, cmn.ChannelSize)
			go filter(edge, in_edge, in_event, new_edge_ch, new_event_ch, out_alert)
			// set the new input channels of the generator
			in_edge = new_edge_ch
			in_event = new_event_ch

		case event, ok := <-in_event:
			if !ok {
				// TODO: Manage the error properly
				fmt.Println("G - !ok in in_event channel")
			}
			// Send the event to the Sink
			// TOCHECK: Send all events to the sink or only some? - filter them?
			out_event <- event
			switch event.Type {
			case cmn.EOF:
				fmt.Println("G - EOF event")
				// end the generator
				break Loop
				/*case cmn.LOG:
				// TODO
				fmt.Println("G: LOG - event")
				// TODO-FUTURE: case: Reconnection case - use this channel?
				*/
			}
		}
	}

	fmt.Println("G finished")
	fmt.Println("G - Close ch - out_alert")
	close(out_alert)
	fmt.Println("G - Close ch - out_event")
	close(out_event)
}

func filter(
	edge cmn.Edge,
	in_edge <-chan cmn.Edge,
	in_event <-chan cmn.Event,
	out_edge chan<- cmn.Edge,
	out_event chan<- cmn.Event,
	out_alert chan<- cmn.Alert) {

	// filter id: is the Card unique identifier
	var id string = edge.Number_id
	var msg_id string = "F-[" + id + "]"

	fmt.Println(msg_id + " - creation")

	/*
		// TODO: Revise the goroutines anonimous to have inside this goroutine...
		int_edge := make(chan cmn.Edge, channelSize)
		int_time := make(chan time.Time) // synchronous
		// TOCHECK: Avoid this channel being blocking (?) Does it make sense?
		int_stop := make(chan bool) // synchronous

		go filter_worker(edge, int_edge, int_time, int_stop, out_alerts)
	*/
Loop:
	for {
		select {
		case edge, ok := <-in_edge:
			if !ok {
				// TODO: Manage the error properly
				fmt.Println("F: !ok in in_edge channel")
			}
			cmn.PrintEdge(msg_id+" - edge arrived", edge)
			if edge.Number_id == id {
				// int_edge <- edge
				cmn.PrintEdge(msg_id+" - belonging edge: ", edge)
			} else {
				out_edge <- edge
			}
		case event, ok := <-in_event:
			if !ok {
				// TODO: Manage the error properly
				fmt.Println(msg_id + "- !ok in in_event channel")
			}
			// Send the event to the next
			out_event <- event
			switch event.Type {
			case cmn.EOF:
				fmt.Println(msg_id + " - EOF event")
				// finish the Filter
				break Loop
				// TODO-FUTURE: case: Reconnection case - use this channel?
			}
		}
	}
	fmt.Println(msg_id + " - Filter finished")
	fmt.Println(msg_id + " - Close ch - out_edge")
	close(out_edge)
	fmt.Println(msg_id + " - Close ch - out_event")
	close(out_event)
}

/*
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
*/

func Source(istream string, out_edge chan<- cmn.Edge, out_event chan<- cmn.Event) {

	// input stream file
	file, err := os.Open(istream)
	cmn.CheckError(err)
	defer file.Close()
	cmn.CheckError(err)

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read() // Read and discard the header line
	cmn.CheckError(err)

	var r cmn.Event
	for {

		tx, err := reader.Read()
		if err == io.EOF {
			fmt.Println("Source - End of stream...")
			r.Type = cmn.EOF
			r.E = cmn.Edge{}
			out_event <- r
			break
		}
		cmn.CheckError(err)

		// conversions
		// id
		tx_id_64, err := strconv.ParseInt(tx[0], 10, 32) // 10: base (decimal) & 32: bit-size (int32)
		cmn.CheckError(err)
		tx_id := int32(tx_id_64) // still the type returned is int64 -> convert to int32

		// type
		var tx_type cmn.TxType
		tx_type_64, err := strconv.ParseInt(tx[3], 10, 8) // int8
		cmn.CheckError(err)
		if tx_type_64 < 0 || tx_type_64 > 3 {
			tx_type = cmn.Other
		} else {
			tx_type = cmn.TxType(tx_type_64)
		}

		// start
		tx_start, err := time.Parse(cmn.Time_layout, tx[4])
		cmn.CheckError(err)

		// end
		// Check if tx_end field is empty
		// From: https://pkg.go.dev/time#Time
		// The zero value of type Time is January 1, year 1, 00:00:00.000000000 UTC. As this time
		// is unlikely to come up in practice, the Time.IsZero method gives a simple way of detecting
		// a time that has not been initialized explicitly.
		var tx_end time.Time
		if tx[5] != "" {
			tx_end, err = time.Parse(cmn.Time_layout, tx[5])
			cmn.CheckError(err)
		}

		var tx_amount_32 float32
		if tx[6] != "" {
			tx_amount, err := strconv.ParseFloat(tx[6], 32)
			cmn.CheckError(err)
			tx_amount_32 = float32(tx_amount)
		}

		edge := cmn.Edge{
			Number_id: tx[1],
			ATM_id:    tx[2],
			Tx_id:     tx_id,
			Tx_type:   tx_type,
			Tx_start:  tx_start,
			Tx_end:    tx_end,
			Tx_amount: tx_amount_32,
		}

		cmn.PrintEdgeComplete("Source - ", edge)

		// TODO/TOCHECK:
		// Do a type for the Edges, instead of Edge do a type!?
		// - Differentiate between Tx_start and Tx_end type of TX
		// - Differentiate between the different types of TX (withdrawal/deposit...)
		out_edge <- edge

		// TODO-REMOVE: -- Only for testing/debugging purposes --
		//  Sleep time for debugging to slow down the flux of transactions
		// Leave without this sleep / change it
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Source - Input finished")
	fmt.Println("Source - Close ch - out_edge")
	close(out_edge)
	fmt.Println("Source - Close ch - out_event")
	close(out_event)

}
