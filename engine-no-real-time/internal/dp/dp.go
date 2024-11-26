package dp

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	cmn "pipeline/internal/common"
	"pipeline/internal/connection"

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
	file_fp_1, err := os.Create(cmn.OutDirName + "/alerts.txt")
	cmn.CheckError(err)
	defer file_fp_1.Close()

	// Logging file
	file_log, err := os.Create(cmn.OutDirName + "/out-log.txt")
	cmn.CheckError(err)
	defer file_log.Close()

	// Results files: diefpy format
	// - trace.csv
	// - metrics.csv

	// trace.csv
	file_trace, err := os.Create(cmn.OutDirName + "/trace.csv")
	cmn.CheckError(err)
	defer file_trace.Close()
	// csv writer
	writer_trace := csv.NewWriter(file_trace)
	defer writer_trace.Flush()
	// headers
	headers := []string{"test", "approach", "answer", "time"}
	err = writer_trace.Write(headers)
	cmn.CheckError(err)

	// metrics.csv
	file_metrics, err := os.Create(cmn.OutDirName + "/metrics.csv")
	cmn.CheckError(err)
	defer file_metrics.Close()
	writer_metrics := csv.NewWriter(file_metrics)
	defer writer_metrics.Flush()
	headers = []string{"test", "approach", "tfft", "totaltime", "comp"}
	err = writer_metrics.Write(headers)
	cmn.CheckError(err)

	var timeFirst time.Duration // variable to keep the time to first answer
	var timeLast time.Duration  // keep the time to last answer

Loop:
	for {
		select {
		case alert, ok := <-in_alert:
			if ok {
				t := time.Since(start_time)
				alertCount += 1
				// save the tfft - metrics.csv
				if alertCount == 1 {
					timeFirst = t
				}
				timeLast = t
				cmn.PrintAlertVerbose(alert, t, alertCount)
				cmn.PrintAlertOnFile(alert, t, alertCount, file_fp_1)
				cmn.PrintAlertOnResultsTrace(t, alertCount, writer_trace)
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
				/*case cmn.LOG:
				// TODO-FUTURE
				*/
			}
		}
	}

	cmn.PrintMetricsResults(timeFirst, timeLast, alertCount, writer_metrics)

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
			//cmn.PrintEdge("G - edge_start arrived: ", event.E)
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

	// Session creation - 1 session per filter - to connect to the gdb
	context := context.Background() // TOCHECK: Use a different new ctx per filter or the same in all?
	session := connection.CreateSession(context)
	defer connection.CloseSession(context, session)

	// Worker - Anonymous function
	go func() {
		var msg_id string = "FW-[" + id + "]"
		var subgraph *cmn.Graph // variable to work with the subgraphs of the different cards

		fmt.Println(msg_id + " - creation")

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
				//cmn.PrintEdge(msg_id+"- edge_start arrived: ", event_worker.E)
				subgraph, ok = card_map[event_worker.E.Number_id]
				if !ok {
					// TODO: Manage the error properly
					fmt.Println("FW - not existing entry in map for: ", event_worker.E.Number_id)
				}
				// 1. Check fraud
				fmt.Println(event_worker.E.Number_id, "-------------- CHECKFRAUD()-----------------")
				isFraud, alert := subgraph.CheckFraud(context, session, event_worker.E)
				fmt.Println("----------------------------------------------------")
				if isFraud {
					out_alert <- alert
				}
				//fmt.Println(msg_id + "................... SUBGRAPH ........................")
				// 2. Add to the subgraph
				subgraph.AddEdge(event_worker.E)
				//subgraph.Print()
			case cmn.EdgeEnd:
				//cmn.PrintEdge(msg_id+"- edge_end arrived: ", event_worker.E)
				subgraph, ok = card_map[event_worker.E.Number_id]
				if !ok {
					// TODO: Manage the error properly
					fmt.Println("FW - not existing entry in map for: ", event_worker.E.Number_id)
				}
				subgraph.CompleteEdge(event_worker.E)
				//subgraph.Print()
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
			/*case cmn.LOG:
			// TODO-FUTURE
			*/
			// TODO-FUTURE: case: Reconnection case - use this channel?
		// TODO: Separate in 2 different cases: EdgeStart and EdgeEnd cases ?
		// --> a EdgeEnd should not be able to create an entry on the map
		// for the moment: ASSUMPTION - tx_end can not arrive before tx_start
		case cmn.EdgeStart, cmn.EdgeEnd:
			// check if edge belongs to filter
			_, ok = card_map[event.E.Number_id]
			if ok {
				//cmn.PrintEdge(msg_id+" - belonging edge: ", event.E)
				internal_edge <- event
			} else if len(card_map) < cmn.MaxFilterSize {
				// filter is not full yet, assign this filter to this card
				//cmn.PrintEdge(msg_id+" - new belonging edge: ", event.E)
				card_map[event.E.Number_id] = cmn.NewGraph()
				internal_edge <- event
			} else {
				//cmn.PrintEdge(msg_id+" - NOT belonging edge: ", event.E)
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

// Source: reads edges given by Stream process
func Source(in_stream <-chan cmn.Event, out_event chan<- cmn.Event) {

	txLogFile, err := os.Create(cmn.OutDirName + "/txLog.txt")
	cmn.CheckError(err)
	defer txLogFile.Close()

	for {
		event, ok := <-in_stream
		if !ok {
			// TODO: Manage the error properly
			fmt.Println("Source - !ok in in_stream channel")
		}
		out_event <- event
		if event.Type == cmn.EOF {
			fmt.Println("Source - EOF event")
			break
		} else if event.Type == cmn.EdgeStart || event.Type == cmn.EdgeEnd {
			// Print the incoming tx in the tx record
			cmn.PrintEdgeCompleteToFile("", event.E, txLogFile)
		}
	}

	fmt.Println("Source - Close ch - out_event")
	close(out_event)
	fmt.Println("Source - Finished")
}

// TODO: Read by chunks so that the reading is not the bottleneck
func Stream(istream string, out_stream chan<- cmn.Event) {

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
			r.Type = cmn.EOF
			r.E = cmn.Edge{}
			out_stream <- r
			break
		}
		cmn.CheckError(err)
		r = cmn.ReadEdge(tx)
		out_stream <- r
	}

	fmt.Println("Stream - End of stream...")
	fmt.Println("Stream - Close ch - out_stream")
	close(out_stream)
	fmt.Println("Stream - Finished")
}
