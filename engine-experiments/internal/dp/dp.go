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

func Sink(
	start_time time.Time,
	in_alert <-chan cmn.Alert,
	in_event <-chan cmn.Event,
	endchan chan<- struct{}) {

	// TOCHECK: Take the initial time here or in the main process before the running of the goroutines...
	// start := time.Now()
	var alertCount int

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
				t := time.Since(start_time)
				alertCount += 1
				fmt.Println("Sink - alert!: ", alert)
				cmn.PrintAlertVerbose(alert, t, alertCount)
				cmn.PrintAlertOnFile(alert, t, alertCount, file_fp_1)
			}
		case event, ok := <-in_event:
			if ok {
				// TODO: Print the event and not (only) the edge associated?
				cmn.PrintEventOnFile(event, file_log)
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
	fmt.Println("Sink - Finished")
}

func Generator(
	in_event <-chan cmn.Event,
	out_alert chan<- cmn.Alert,
	out_event chan<- cmn.Event) {

	fmt.Println("G - creation")
Loop:
	for {
		event, ok := <-in_event
		if !ok {
			// TODO: Manage the error properly
			fmt.Println("G - !ok in in_event channel")
		}
		switch event.Type {
		case cmn.EOF:
			fmt.Println("G - EOF event")
			out_event <- event
			// end the generator
			break Loop
			/*case cmn.LOG:
			// TODO
			fmt.Println("G: LOG - event")
			// TODO-FUTURE: case: Reconnection case - use this channel?
			*/
		case cmn.EdgeEnd:
			// TODO: decide how to manage better?
			log.Fatalf("Error: edge_end arrived before edge_start")
		case cmn.EdgeStart:
			cmn.PrintEdge("G - edge_start arrived: ", event.E)
			// spawn a filter
			new_event_ch := make(chan cmn.Event, cmn.ChannelSize)
			go filter(event, in_event, new_event_ch, out_alert)
			// set the new input channels of the generator
			in_event = new_event_ch
		}
	}

	fmt.Println("G - Close ch - out_alert")
	close(out_alert)
	fmt.Println("G - Close ch - out_event")
	close(out_event)
	fmt.Println("G finished")
}

// TODO: pass a counter to use as filter-id instead of the id of the card that spawns it?
func filter(
	event cmn.Event,
	in_event <-chan cmn.Event,
	out_event chan<- cmn.Event,
	out_alert chan<- cmn.Alert) {

	var edge cmn.Edge = event.E
	var id string = edge.Number_id
	var msg_id string = "F-[" + id + "]"
	fmt.Println(msg_id + " - creation")
	// hash table to index card ids to card subgraphs
	// key: card id
	// value: pointer to Graph
	// NOTE: maps are inherently dynamic in size. -> control the desired
	// max size by ourselves
	/*
		FILTER:
			- Reads to check the existance of an entry on the map
			- Creates the entries on the map (in the corresponding cases)
		WORKER:
			- Modifies the entries on the map once they are created
			(Filter does not modify values after the creation of the entry)

		// Conclusion: Safe to do it with a single map and without mutex, since
		there can not be concurrent writes on the same map entries. Filter writes
		on the creation and then it is only the worker who writes on that entry after
		the creation of the entry by the filter.
	*/
	var card_map map[string]*cmn.Graph = make(map[string]*cmn.Graph)
	card_map[edge.Number_id] = cmn.NewGraph() // first entry creation

	// internal_edge channel between Filter and Worker - only pass events of type Edge (EdgeStart or EdgeEnd)
	internal_edge := make(chan cmn.Event, cmn.ChannelSize)
	// synchronization channel between Filter and Worker, to let Filter know whenever Worker is done
	endchan := make(chan struct{})

	// Worker - Anonymous function
	go func() {
		var msg_id string = "FW-[" + id + "]"
		var subgraph *cmn.Graph // variable to work with the subgraphs of the different cards

		fmt.Println(msg_id + " - creation")
		cmn.PrintEdge(msg_id+" - initial edge:", edge)

		subgraph, ok := card_map[edge.Number_id]
		if !ok {
			// TODO: Manage the error properly
			fmt.Println("FW - not existing entry in map for: ", edge.Number_id)
		}

		subgraph.AddEdge(edge)
		// subgraph.PrintIds()

		// this goroutine dies alone after its father closes the internal_edge channel
		// (it is the only process with which it has communication / is connected)
	Worker_Loop:
		for {
			event_worker, ok := <-internal_edge
			if !ok {
				// TODO: Check what to do here better
				fmt.Println(msg_id + "- closed internal_edge channel")
				break Worker_Loop
			}

			switch event_worker.Type {
			case cmn.EOF:
				// finish the worker
				endchan <- struct{}{}
				break Worker_Loop
			case cmn.EdgeStart:
				// start edge
				cmn.PrintEdge(msg_id+"- edge_start arrived: ", event_worker.E)
				subgraph, ok = card_map[event_worker.E.Number_id]
				if !ok {
					// TODO: Manage the error properly
					fmt.Println("FW - not existing entry in map for: ", event_worker.E.Number_id)
				}
				// 1. Check fraud
				fmt.Println("-------------- CHECKFRAUD()-----------------")
				isFraud, alert := subgraph.CheckFraud(event_worker.E)
				fmt.Println("--------------------------------------------")
				if isFraud {
					out_alert <- alert
				}
				//fmt.Println(msg_id + "................... SUBGRAPH ........................")
				// 2. Add to the subgraph
				subgraph.AddEdge(event_worker.E)
				subgraph.Print()
			case cmn.EdgeEnd:
				cmn.PrintEdge(msg_id+"- edge_end arrived: ", event_worker.E)
				subgraph, ok = card_map[event_worker.E.Number_id]
				if !ok {
					// TODO: Manage the error properly
					fmt.Println("FW - not existing entry in map for: ", event_worker.E.Number_id)
				}
				subgraph.CompleteEdge(event_worker.E)
				subgraph.Print()
			}
		}
		fmt.Println(msg_id + " - Filter worker finished")
	}() // () here to not only define it but also run it

Loop:
	for {
		event, ok := <-in_event
		if !ok {
			// TODO: Manage the error properly
			fmt.Println(msg_id + "- !ok in in_event channel")
		}
		switch event.Type {
		case cmn.EOF:
			fmt.Println(msg_id + " - EOF event")
			// finish the Filter
			// pass the EOF event to the worker & wait until its worker is done
			internal_edge <- event
			<-endchan
			// pass the finish event to next process
			out_event <- event // TOCHECK: This before worker is done or here?
			break Loop
			// TODO-FUTURE: case: Reconnection case - use this channel?
		// TODO: Separate in 2 different cases: EdgeStart and EdgeEnd cases ?
		// --> a EdgeEnd should not be able to create an entry on the map
		// for the moment: ASSUMPTION - tx_end can not arrive before tx_start
		case cmn.EdgeStart, cmn.EdgeEnd:
			//fmt.Println(">>>>>>>>>>>>> num-cards: ", len(card_map))
			// check if edge belongs to filter
			_, ok = card_map[event.E.Number_id]
			if ok {
				cmn.PrintEdge(msg_id+" - belonging edge: ", event.E)
				internal_edge <- event
			} else if len(card_map) < cmn.MaxFilterSize {
				// filter is not full yet, assign this filter to this card
				cmn.PrintEdge(msg_id+" - new belonging edge: ", event.E)
				card_map[event.E.Number_id] = cmn.NewGraph()
				internal_edge <- event
			} else {
				cmn.PrintEdge(msg_id+" - NOT belonging edge: ", event.E)
				out_event <- event
			}
		}
	}

	fmt.Println(msg_id + " - Close ch - internal_edge")
	close(internal_edge)
	fmt.Println(msg_id + " - Close ch - out_event")
	close(out_event)
	fmt.Println(msg_id + " - Filter finished")
}

func Source(istream string, out_event chan<- cmn.Event) {

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
			r.Type = cmn.EdgeEnd
		} else {
			r.Type = cmn.EdgeStart // tx_end field is empty
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

		//cmn.PrintEdgeComplete("Source - ", edge)

		// TODO/TOCHECK:
		// Do a type for the Edges, instead of Edge do a type!?
		// - Differentiate between Tx_start and Tx_end type of TX
		// - Differentiate between the different types of TX (withdrawal/deposit...)
		r.E = edge
		out_event <- r

		// TODO-REMOVE: -- Only for testing/debugging purposes --
		//  Sleep time for debugging to slow down the flux of transactions
		// Leave without this sleep / change it
		//time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Source - Close ch - out_event")
	close(out_event)
	fmt.Println("Source - Finished")
}
