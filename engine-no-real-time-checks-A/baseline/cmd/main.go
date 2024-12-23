/*
Entry point of the program
*/

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	cmn "pipeline/internal/common"
	"pipeline/internal/connection"
	"pipeline/internal/dp"
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

	// creation of needed channels
	// real-time input stream channel
	stream_ch := make(chan cmn.Event, cmn.ChannelSize)

	// event channel
	//event_ch := make(chan cmn.Event, cmn.ChannelSize)
	// alerts channel
	//alert_ch := make(chan cmn.Alert, cmn.ChannelSize)
	// out_event_ch channel: direct event channel between Generator and Sink.
	// --> all kinds of events except the alerts
	//out_event_ch := make(chan cmn.Event, cmn.ChannelSize)
	// Ending channel
	//endchan := make(chan struct{})
	start := time.Now()

	// Session creation
	session := connection.CreateSession(ctx)
	defer connection.CloseSession(ctx, session)

	// launch Stream goroutines - to provide the input in real-time
	go dp.Stream(cmn.StreamFileName, stream_ch)
	// launch Source, Generator and Sink goroutines
	// hash table to index card ids to card subgraphs
	// - 1 to map each belonging card to its corresponding subgraph	(cardSubgraph)
	var cardSubgraph map[string]*cmn.Graph = make(map[string]*cmn.Graph)
	var subgraph *cmn.Graph

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

	var finish bool = false

	for !finish {
		event_stream, ok := <-stream_ch
		if !ok {
			// TODO: Check what to do here better
			fmt.Println("Closed stream_ch channel")
			break
		}

		switch event_stream.Type {
		case cmn.EOF:
			fmt.Println("EOF event")
			finish = true
		case cmn.EdgeStart:
			// check if card exists in cardSubgraph map & create entry if it does not exist
			subgraph, ok = cardSubgraph[event_stream.E.Number_id]
			if !ok {
				// card does not exist -> create entry for the new card in the cardSubgraph
				// and add the new edge
				cardSubgraph[event_stream.E.Number_id] = cmn.NewGraph()
				subgraph, ok = cardSubgraph[event_stream.E.Number_id]
				if !ok {
					fmt.Println("Not existing entry in map for: ", event_stream.E.Number_id)
				}
				// add to the subgraph
				subgraph.AddEdge(event_stream.E)
			} else {
				// card already exists, therefore, at least an edge on the subgraph
				// check fraud
				//fmt.Println(event_worker.E.Number_id, "-------------- CHECKFRAUD()-----------------")
				isFraud, alert := subgraph.CheckFraud(ctx, session, event_stream.E)
				//fmt.Println("----------------------------------------------------")
				if isFraud {
					t := time.Since(start)
					alertCount += 1
					// save the tfft - metrics.csv
					if alertCount == 1 {
						timeFirst = t
					}
					timeLast = t
					//cmn.PrintAlertVerbose(alert, t, alertCount)
					cmn.PrintAlertOnFile(alert, fileAlerts)
					cmn.PrintAlertOnResultsTrace(t, alertCount, writer_trace)
				}
				// set as new head of the subgraph (only save the last edge)
				subgraph.NewHead(event_stream.E)
			}
		case cmn.EdgeEnd:
			//cmn.PrintEdge(msg_id+"- edge_end arrived: ", event_worker.E)
			subgraph, ok = cardSubgraph[event_stream.E.Number_id]
			if !ok {
				// TODO: Manage the error properly
				fmt.Println("Edge end has not existing entry in map for: ", event_stream.E.Number_id)
				log.Println("Warning: AddEdge -> a tx-end was tryied to be added on a empty subgraph", event_stream.E.Number_id)
				// NOTE: THIS SHOULD NOT BE DONE HERE - tx_start should arrive before tx_end
				// Warn - anyway create the subgraph for this edge
				cardSubgraph[event_stream.E.Number_id] = cmn.NewGraph()
				subgraph, ok = cardSubgraph[event_stream.E.Number_id]
				if !ok {
					// TODO: Manage the error properly
					fmt.Println("Not existing entry in map for: ", event_stream.E.Number_id)
				}
				subgraph.AddEdge(event_stream.E)
			} else {
				subgraph.CompleteEdge(event_stream.E)
			}
		}
	}

	cmn.PrintMetricsResults(timeFirst, timeLast, alertCount, writer_metrics)

	t := time.Since(start)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")
	fmt.Println("Finish Program")

	// TODO: finish connection to static gdb
	connection.CloseConnection(ctx)

}
