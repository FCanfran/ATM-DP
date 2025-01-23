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
	headers := []string{"test", "approach", "answer", "time", "responseTime", "isPositive"}
	err = writer_trace.Write(headers)
	cmn.CheckError(err)

	// metrics.csv
	file_metrics, err := os.Create(cmn.OutDirName + "/metrics.csv")
	cmn.CheckError(err)
	defer file_metrics.Close()
	writer_metrics := csv.NewWriter(file_metrics)
	defer writer_metrics.Flush()
	headers = []string{"test", "approach", "tfft", "totaltime", "mrt", "checks", "alerts"}
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
				responseTime := t - check.LastEventTimestamp
				// Record verbose on a file only the alerts
				if check.IsPositive {
					alertCount += 1
					cmn.PrintAlertOnFileVerbose(check, responseTime, alertCount, fileAlerts)
				}
				cmn.PrintCheckOnResultsTrace(t, responseTime, checkCount, check.IsPositive, writer_trace)
			}
		case event, ok := <-in_event:
			if ok {
				// TODO: Print the event and not (only) the edge associated?
				cmn.PrintEventOnFile(event, file_log)
			}
			switch event.Type {
			case cmn.EOF:
				// finish the Sink
				break Loop
			}
		}
	}

	cmn.PrintMetricsResults(timeFirst, timeLast, checkCount, alertCount, writer_metrics)

	fmt.Println("numChecks: ", checkCount)
	fmt.Println("numAlerts: ", alertCount)

	endchan <- struct{}{}
}

func Generator(
	in_event <-chan cmn.Event,
	out_check chan<- cmn.CheckResult,
	out_event chan<- cmn.Event) {

Loop:
	for {
		event, ok := <-in_event
		if !ok {
			fmt.Println("G - !ok in in_event channel")
		}
		switch event.Type {
		case cmn.EOF:
			out_event <- event
			// end the generator
			break Loop
		case cmn.EdgeEnd:
			log.Fatalf("Error-Generator: edge_end arrived before edge_start")
		case cmn.EdgeStart:
			// spawn a filter
			new_event_ch := make(chan cmn.Event, cmn.ChannelSize)
			go filter(event, in_event, new_event_ch, out_check)
			// set the new input channels of the generator
			in_event = new_event_ch
		}
	}

	close(out_check)
	close(out_event)
}

func filter(
	event cmn.Event,
	in_event <-chan cmn.Event,
	out_event chan<- cmn.Event,
	out_check chan<- cmn.CheckResult) {

	var edge cmn.Edge = event.E
	var id string = edge.Number_id
	var msg_id string = "F-[" + id + "]"

	var cardList map[string]bool = make(map[string]bool)
	var cardSubgraph map[string]*cmn.Graph = make(map[string]*cmn.Graph)

	cardList[edge.Number_id] = true

	// internal_edge channel between Filter and Worker - only pass events of type Edge (EdgeStart or EdgeEnd)
	internal_edge := make(chan cmn.Event, cmn.ChannelSize)
	// synchronization channel between Filter and Worker, to let Filter know whenever Worker is done
	endchan := make(chan struct{})

	// Session creation - 1 session per filter - to connect to the gdb
	context := context.Background()
	session := connection.CreateSession(context)
	defer connection.CloseSession(context, session)

	// Worker - Anonymous function
	go func() {
		var subgraph *cmn.Graph // variable to work with the subgraphs of the different cards
		cardSubgraph[edge.Number_id] = cmn.NewGraph()
		subgraph, ok := cardSubgraph[edge.Number_id]
		if !ok {
			fmt.Println("FW - not existing entry in map for: ", edge.Number_id)
		}

		subgraph.AddEdge(edge)
		// this goroutine dies alone after its father closes the internal_edge channel
		// (it is the only process with which it has communication / is connected)
	Worker_Loop:
		for {
			event_worker, ok := <-internal_edge
			if !ok {
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
						fmt.Println("FW - not existing entry in map for: ", event_worker.E.Number_id)
					}
					// add to the subgraph
					subgraph.AddEdge(event_worker.E)
				} else {
					// card already exists, therefore, at least an edge on the subgraph
					result := subgraph.CheckFraud(context, session, event_worker.E)
					result.LastEventTimestamp = event_worker.Timestamp
					out_check <- result
					// set as new head of the subgraph (only save the last edge)
					subgraph.NewHead(event_worker.E)
				}
			case cmn.EdgeEnd:
				subgraph, ok = cardSubgraph[event_worker.E.Number_id]
				if !ok {
					fmt.Println("FW - edge end has not existing entry in map for: ", event_worker.E.Number_id)
					log.Println("Warning: AddEdge -> a tx-end was tryied to be added on a empty subgraph", event_worker.E.Number_id)
					cardSubgraph[event_worker.E.Number_id] = cmn.NewGraph()
					subgraph, ok = cardSubgraph[event_worker.E.Number_id]
					if !ok {
						fmt.Println("FW - not existing entry in map for: ", event_worker.E.Number_id)
					}
					subgraph.AddEdge(event_worker.E)
				} else {
					subgraph.CompleteEdge(event_worker.E)
				}
			}
		}
	}()

Loop:
	for {
		event, ok := <-in_event
		if !ok {
			fmt.Println(msg_id + "- !ok in in_event channel")
		}
		switch event.Type {
		case cmn.EOF:
			// finish the Filter
			// pass the EOF event to the worker & wait until its worker is done
			internal_edge <- event
			<-endchan
			// pass the finish event to next process
			out_event <- event
			break Loop
		case cmn.EdgeStart, cmn.EdgeEnd:
			// check if edge belongs to filter - true if exists and zero-value (false) otherwise
			if cardList[event.E.Number_id] {
				internal_edge <- event
			} else if len(cardList) < cmn.MaxFilterSize {
				// filter is not full yet, assign this filter to this card
				cardList[event.E.Number_id] = true
				internal_edge <- event
			} else {
				out_event <- event
			}
		}
	}

	close(internal_edge)
	close(out_event)
}

// Source: reads edges given by Stream process
func Source(start_time time.Time, in_stream <-chan cmn.Event, out_event chan<- cmn.Event) {

	txLogFile, err := os.Create(cmn.OutDirName + "/txLog.txt")
	cmn.CheckError(err)
	defer txLogFile.Close()

	for {
		event, ok := <-in_stream
		if !ok {
			fmt.Println("Source - !ok in in_stream channel")
		}
		// get internal system event timestamp - to mark/simulate when the event arrived to the system
		t := time.Since(start_time)
		event.Timestamp = t
		out_event <- event
		if event.Type == cmn.EOF {
			break
		} else if event.Type == cmn.EdgeStart || event.Type == cmn.EdgeEnd {
			cmn.PrintEdgeCompleteToFile("", event.E, txLogFile)
		}
	}

	close(out_event)
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
			out_stream <- event
			rows++
		}
	}

	fmt.Println("rows: ------------> ", rows)

	// send EOF event
	event.Type = cmn.EOF
	event.E = cmn.Edge{}
	out_stream <- event

	close(out_stream)
}
