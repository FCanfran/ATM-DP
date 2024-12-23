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
	in_check <-chan cmn.CheckResult,
	in_event <-chan cmn.Event,
	endchan chan<- struct{}) {

	var checkCount int
	var alertCount int

	fileAlerts, err := os.Create(cmn.OutDirName + "/alerts.txt")
	cmn.CheckError(err)
	defer fileAlerts.Close()

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
	headers := []string{"test", "approach", "answer", "time", "responseTimeSink", "responseTimeFilter", "rtDiff", "isPositive"}
	err = writer_trace.Write(headers)
	cmn.CheckError(err)

	// metrics.csv
	file_metrics, err := os.Create(cmn.OutDirName + "/metrics.csv")
	cmn.CheckError(err)
	defer file_metrics.Close()
	writer_metrics := csv.NewWriter(file_metrics)
	defer writer_metrics.Flush()
	headers = []string{"test", "approach", "tfft", "totaltime", "mrt", "comp"}
	err = writer_metrics.Write(headers)
	cmn.CheckError(err)

	var timeFirst time.Duration // variable to keep the time to first answer
	var timeLast time.Duration  // keep the time to last answer

Loop:
	for {
		select {
		case check, ok := <-in_check:
			if ok {
				t := time.Since(start_time)
				checkCount += 1
				// save the tfft - metrics.csv
				if checkCount == 1 {
					timeFirst = t
				}
				timeLast = t
				// calculate response time
				responseTimeSink := t - check.LastEventTimestamp
				// Record verbose on a file only the alerts
				if check.IsPositive {
					alertCount += 1
					cmn.PrintAlertOnFileVerbose(check, responseTimeSink, alertCount, fileAlerts)
				}

				// print the difference between the responseTime measured on Filter and on Sink
				diffResponseTime := responseTimeSink - check.ResponseTime
				cmn.PrintCheckOnResultsTrace(t, responseTimeSink, checkCount, check.IsPositive, writer_trace, check.ResponseTime, diffResponseTime)
			}
		case event, ok := <-in_event:
			if ok {
				// TODO: Print the event and not (only) the edge associated?
				cmn.PrintEventOnFile(event, file_log)
			}
			switch event.Type {
			case cmn.EOF:
				//fmt.Println("Sink - EOF event")
				// finish the Sink
				break Loop
				/*case cmn.LOG:
				// TODO-FUTURE
				*/
			}
		}
	}

	cmn.PrintMetricsResults(timeFirst, timeLast, checkCount, writer_metrics)

	endchan <- struct{}{}
}

func Generator(
	start_time time.Time,
	in_event <-chan cmn.Event,
	out_check chan<- cmn.CheckResult,
	out_event chan<- cmn.Event) {

	//fmt.Println("G - creation")
Loop:
	for {
		event, ok := <-in_event
		if !ok {
			// TODO: Manage the error properly
			fmt.Println("G - !ok in in_event channel")
		}
		switch event.Type {
		case cmn.EOF:
			//fmt.Println("G - EOF event")
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
			log.Fatalf("Error-Generator: edge_end arrived before edge_start")
		case cmn.EdgeStart:
			//cmn.PrintEdge("G - edge_start arrived: ", event.E)
			// spawn a filter
			new_event_ch := make(chan cmn.Event, cmn.ChannelSize)
			go filter(start_time, event, in_event, new_event_ch, out_check)
			// set the new input channels of the generator
			in_event = new_event_ch
		}
	}

	//fmt.Println("G - Close ch - out_alert")
	close(out_check)
	//fmt.Println("G - Close ch - out_event")
	close(out_event)
	//fmt.Println("G finished")
}

func filter(
	start_time time.Time,
	event cmn.Event,
	in_event <-chan cmn.Event,
	out_event chan<- cmn.Event,
	out_check chan<- cmn.CheckResult) {

	var edge cmn.Edge = event.E
	var id string = edge.Number_id
	var msg_id string = "F-[" + id + "]"
	//fmt.Println(msg_id + " - creation")
	// hash table to index card ids to card subgraphs
	// 2 hash tables (to avoid race conditions in concurrent access by filter & worker)
	// - 1 to control the belonging cards to the filter 			(cardList)		-> only access by filter
	// - 1 to map each belonging card to its corresponding subgraph	(cardSubgraph)	-> only access by worker
	// NOTE: maps are inherently dynamic in size. -> control the desired
	// max size by ourselves
	var cardList map[string]bool = make(map[string]bool)
	var cardSubgraph map[string]*cmn.Graph = make(map[string]*cmn.Graph)

	cardList[edge.Number_id] = true

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
		//var msg_id string = "FW-[" + id + "]"
		var subgraph *cmn.Graph // variable to work with the subgraphs of the different cards
		//fmt.Println(msg_id + " - creation")

		cardSubgraph[edge.Number_id] = cmn.NewGraph()
		subgraph, ok := cardSubgraph[edge.Number_id]
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
				//fmt.Println(msg_id + "- closed internal_edge channel")
				break Worker_Loop
			}

			switch event_worker.Type {
			case cmn.EOF:
				// finish the worker
				endchan <- struct{}{}
				break Worker_Loop
			case cmn.EdgeStart:
				// check if card exists in cardSubgraph map & create entry if it does not exist
				subgraph, ok = cardSubgraph[event_worker.E.Number_id]
				if !ok {
					// card does not exist -> create entry for the new card in the cardSubgraph
					// and add the new edge
					cardSubgraph[event_worker.E.Number_id] = cmn.NewGraph()
					subgraph, ok = cardSubgraph[event_worker.E.Number_id]
					if !ok {
						// TODO: Manage the error properly
						fmt.Println("FW - not existing entry in map for: ", event_worker.E.Number_id)
					}
					// add to the subgraph
					subgraph.AddEdge(event_worker.E)
				} else {
					// card already exists, therefore, at least an edge on the subgraph
					result := subgraph.CheckFraud(context, session, event_worker.E)
					t := time.Since(start_time)
					result.ResponseTime = t - event_worker.Timestamp
					result.LastEventTimestamp = event_worker.Timestamp
					out_check <- result
					// set as new head of the subgraph (only save the last edge)
					subgraph.NewHead(event_worker.E)
				}
			case cmn.EdgeEnd:
				//cmn.PrintEdge(msg_id+"- edge_end arrived: ", event_worker.E)
				subgraph, ok = cardSubgraph[event_worker.E.Number_id]
				if !ok {
					// TODO: Manage the error properly
					fmt.Println("FW - edge end has not existing entry in map for: ", event_worker.E.Number_id)
					log.Println("Warning: AddEdge -> a tx-end was tryied to be added on a empty subgraph", event_worker.E.Number_id)
					// NOTE: THIS SHOULD NOT BE DONE HERE - tx_start should arrive before tx_end
					// Warn - anyway create the subgraph for this edge
					cardSubgraph[event_worker.E.Number_id] = cmn.NewGraph()
					subgraph, ok = cardSubgraph[event_worker.E.Number_id]
					if !ok {
						// TODO: Manage the error properly
						fmt.Println("FW - not existing entry in map for: ", event_worker.E.Number_id)
					}
					subgraph.AddEdge(event_worker.E)
				} else {
					subgraph.CompleteEdge(event_worker.E)
				}
			}
		}
		//fmt.Println(msg_id + " - Filter worker finished")
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
			//fmt.Println(msg_id + " - EOF event")
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
			// check if edge belongs to filter - true if exists and zero-value (false) otherwise
			if cardList[event.E.Number_id] {
				//cmn.PrintEdge(msg_id+" - belonging edge: ", event.E)
				internal_edge <- event
			} else if len(cardList) < cmn.MaxFilterSize {
				// filter is not full yet, assign this filter to this card
				//cmn.PrintEdge(msg_id+" - new belonging edge: ", event.E)
				cardList[event.E.Number_id] = true
				internal_edge <- event
			} else {
				//cmn.PrintEdge(msg_id+" - NOT belonging edge: ", event.E)
				out_event <- event
			}
		}
	}

	//fmt.Println(msg_id + " - Close ch - internal_edge")
	close(internal_edge)
	//fmt.Println(msg_id + " - Close ch - out_event")
	close(out_event)
	//fmt.Println(msg_id + " - Filter finished")
}

// Source: reads edges given by Stream process
func Source(start_time time.Time, in_stream <-chan cmn.Event, out_event chan<- cmn.Event) {

	txLogFile, err := os.Create(cmn.OutDirName + "/txLog.txt")
	cmn.CheckError(err)
	defer txLogFile.Close()

	for {
		event, ok := <-in_stream
		if !ok {
			// TODO: Manage the error properly
			fmt.Println("Source - !ok in in_stream channel")
		}
		// get internal system event timestamp - to mark/simulate when the event arrived to the system
		t := time.Since(start_time)
		event.Timestamp = t
		out_event <- event
		if event.Type == cmn.EOF {
			//fmt.Println("Source - EOF event")
			break
		} else if event.Type == cmn.EdgeStart || event.Type == cmn.EdgeEnd {
			// Print the incoming tx in the tx record
			cmn.PrintEdgeCompleteToFile("", event.E, txLogFile)
		}
	}

	//fmt.Println("Source - Close ch - out_event")
	close(out_event)
	//fmt.Println("Source - Finished")
}

func Stream(istream string, out_stream chan<- cmn.Event) {

	// channel of chunks - slices of rows
	chunk_ch := make(chan [][]string)

	// worker to do background reading
	go func() {

		file, err := os.Open(istream)
		cmn.CheckError(err)
		defer file.Close()
		cmn.CheckError(err)

		// csv reader
		reader := csv.NewReader(bufio.NewReader(file))
		_, err = reader.Read() // Read and discard the header line
		cmn.CheckError(err)

		var rows [][]string
		i := 0
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			cmn.CheckError(err)
			rows = append(rows, row)
			i++
			if i == cmn.ChunkSize {
				chunk_ch <- rows
				rows = nil // clear the rows holder
				i = 0
			}
		}

		// send the remaining rows if there are
		if len(rows) > 0 {
			chunk_ch <- rows
		}
		close(chunk_ch)
	}()

	var event cmn.Event
	rows := 0
	for chunk := range chunk_ch {
		for _, row := range chunk {
			event = cmn.ReadEdge(row) // converting to corresp. types and creating edge event
			//cmn.PrintEdgeComplete("", event.E)
			out_stream <- event
			rows++
		}
	}

	fmt.Println("rows: ------------> ", rows)

	// send EOF event
	event.Type = cmn.EOF
	event.E = cmn.Edge{}
	out_stream <- event

	//fmt.Println("Stream - End of stream...")
	//fmt.Println("Stream - Close ch - out_stream")
	close(out_stream)
	//fmt.Println("Stream - Finished")
}
