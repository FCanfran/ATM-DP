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

type in_comm struct {
	// read-only channels
	Edge          <-chan cmn.Edge // Edges channel
	Front_Channel <-chan in_comm  // in_comm channel
}

func generator(in <-chan cmn.Edge, log_ch chan cmn.Edge) {
	fmt.Println("G - creation")
	// Generate input channels
	alerts := make(chan cmn.Alert, channelSize)
	// Note: by default channels created by "make" create bidirectional
	// channels by default. To make it receive-only channel, we perform
	// the type conversion as shown
	tmp := make(chan in_comm, channelSize)
	var front_channels <-chan in_comm // read only channel
	front_channels = tmp

	// TOCHECK: Create results output files: one for each kind of fraud pattern (?)
	// TODO: For the moment only 1 kind of pattern
	file_fp_1, err := os.Create("../output/outPattern1.txt")
	cmn.CheckError(err)
	defer file_fp_1.Close()

	file_log, err := os.Create("../output/log.txt")
	cmn.CheckError(err)
	defer file_log.Close()

	for {
		select {
		case edge := <-in:
			//cmn.PrintEdge("G - edge arrived: ", edge)
			// spawn a filter
			// - input channels: the input channels of the generator:
			//					* in - Edge
			// 					* front_channels - Front
			// - output channels:
			// 					* new_edge_ch - Edge (new)
			// 					* alerts - Alerts (same)
			//					* new_front_ch - Front (new)
			// creation of new bidirectional needed channels
			new_edge_ch := make(chan cmn.Edge, channelSize)
			new_front_ch := make(chan in_comm, channelSize)
			go filter(edge, in, front_channels, new_edge_ch, alerts, log_ch, new_front_ch)
			// set the new input channels of the generator
			in = new_edge_ch
			front_channels = new_front_ch
		case alert := <-alerts:
			fmt.Println("G - alert!: ", alert)
			cmn.PrintAlertVerbose(alert)
			cmn.PrintAlertOnFile(alert, file_fp_1)
		case event := <-log_ch:
			cmn.PrinteEventOnFile(event, file_log)
		case input := <-front_channels:
			// Reconnection of the pipeline (case of a filter having died)
			fmt.Println("G - reconnection")
			in = input.Edge
			front_channels = input.Front_Channel
		}
	}
}

// FUTURE: Unused channels
// in_front, out_front - for the future filter's lifetime management - to be able
// to kill a filter and do the reconnection with the pipeline properly
// out_log: to register logging messages that are sent to the generator
func filter(edge cmn.Edge, in_edge <-chan cmn.Edge, in_front <-chan in_comm,
	out_edge chan<- cmn.Edge, out_alert chan<- cmn.Alert, out_log_ch chan<- cmn.Edge, out_front chan<- in_comm) {

	// filter id: is the Card unique identifier
	var id string = edge.Number_id

	fmt.Println("F ", id, "- creation")

	int_edge := make(chan cmn.Edge, channelSize)
	int_time := make(chan time.Time) // synchronous
	// TOCHECK: Avoid this channel being blocking (?) Does it make sense?
	int_stop := make(chan bool) // synchronous

	go filter_worker(edge, int_edge, int_time, int_stop, out_alert, out_log_ch)

	for {
		select {
		case edge := <-in_edge:
			// cmn.PrintEdge("F - edge arrived: ", edge)
			if edge.Number_id == id {
				int_edge <- edge
			} else {
				out_edge <- edge
				// -------------------------------------------------------------------------------------------------- //
				// FUTURE: So far, assuming that the we have a single infinite time window - no management of filter's lifetime
				/*
					// TODO: Gestion del tiempo de vida del filtro con incoming timestamp de los edges que van pasando
					// TOCHECK: Tx_end instead of Tx_start: tx come ordered by tx_end, which is the moment when the tx ended
					// which is when it arrived to our system -> closer to the current real time!
					int_time <- edge.Tx_end
					// TODO: avoid this signal (stop) being synchronous! -
					// allow worker to tell at any moment to stop! instead of blocking

					if stop := <-int_stop; stop {
						// kill the filter and pass the front channel to do the reconnection
						out_front <- in_comm{in_edge, in_front}
						fmt.Println("F ", id, " - kill")
						return // finish filter
					}
				*/
				// -------------------------------------------------------------------------------------------------- //

			}
		case input := <-in_front:
			fmt.Println("F ", id, " - reconnection")
			// a previous filter died, reconnect pipeline
			in_edge = input.Edge
			in_front = input.Front_Channel
		}
	}
}

// FUTURE: Unused channels - for the future filter's lifetime management - to be able
// to kill a filter and do the reconnection with the pipeline properly
func filter_worker(initial_edge cmn.Edge, int_edge <-chan cmn.Edge, int_time <-chan time.Time, int_stop chan<- bool,
	out_alert chan<- cmn.Alert, out_log_ch chan<- cmn.Edge) {

	//cmn.PrintEdge("FW creation - edge arrived: ", initial_edge)
	var subgraph *cmn.Graph = cmn.NewGraph() // Explicit declaration
	subgraph.AddAtEnd(initial_edge)
	subgraph.PrintIds()

	// TODO: this goroutine dies alone after its father (the filter) dies?
	// -> it is the only process with which it has communication / is connected
	for {
		select {
		case new_edge := <-int_edge:
			// FUTURE WORK: for multiple window support - so far: Single window support (do not need this)
			// -------------------------------------------------------------------------------------------------- //
			// TODO: Check the order of these operations
			// NOTE: update the subgraph wrt the timestamp of this new edge
			// first: update the subgraph wrt the timestamp of this new edge and
			// second: add the new edge
			// TOCHECK: Tx_end instead of Tx_start: tx come ordered by tx_end, which is the moment when the tx ended
			// which is when it arrived to our system -> closer to the current real time!
			//subgraph.Update(new_edge.Tx_end)
			// -------------------------------------------------------------------------------------------------- //
			// NOTE: New -> with 2 edges per tx
			// 1. Identify if it is start or end edge
			isStart := new_edge.IsStart()
			if isStart {
				// start edge
				// 1. Check fraud
				// 2. Add to the subgraph
			}
			// Add to the subgraph
			subgraph.AddEdge(new_edge, isStart)
			/*
				// TODO: How to do when the new edge produces fraud pattern? - add/dont add to the volatile subgraph?
				isFraud, fraudSubgraph, anomalousEdge := subgraph.CheckFraud(new_edge)
				if isFraud {
					// TODO: Create & propagate fraud pattern alert (alert channel)
					fmt.Println("FW - Positive Fraud pattern")
					fraud1Alert := cmn.Alert{
						Label:         "1",
						Info:          "fraud pattern",
						Subgraph:      *fraudSubgraph,
						AnomalousEdge: anomalousEdge,
					}
					out_alert <- fraud1Alert

				} else {
					fmt.Println("FW - Negative Fraud pattern")
					subgraph.AddAtEnd(new_edge)
				}
				subgraph.PrintIds()
			*/
			// -------------------------------------------------------------------------------------------------- //
			// FUTURE: So far, assuming that the we have a single infinite time window - no management of filter's lifetime
			/*
				case new_time := <-int_time:
					// 1. Filter Timeout check: test if the filter has to die (with the last edge of the volatile
					// subgraph), in that case send stop signal to the (father) filter
					if subgraph.CheckFilterTimeout(new_time) {
						int_stop <- true
					} else {
						int_stop <- false // TODO: CHECK IF THIS CAN COME HERE OR WE HAVE TO WAIT TO DO IT AT
						// THE END
						// filter is not killed but we need to update the subgraph to (possibly) eliminate the
						// outdated edges on the volatile subgraph
						subgraph.Update(new_time)
						subgraph.PrintIds()
						// int_stop <- false
					}
			*/
			// -------------------------------------------------------------------------------------------------- //
		}

	}
}

func Start(istream string) {

	// Creation of edges channel to pass from the read input to the pipeline
	edges_input := make(chan cmn.Edge, channelSize)

	// Log channel: to register all the events coming to the engine. Bidirectional.
	// Registering in the sink (generator for the moment)
	// TOCHECK: For the moment only edges through it
	log_ch := make(chan cmn.Edge, channelSize)

	go generator(edges_input, log_ch)

	file, err := os.Open(istream)
	cmn.CheckError(err)
	// closes the file after read from it no matter if there is error or not
	defer file.Close()
	cmn.CheckError(err)

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))

	// Read and discard the header line
	_, err = reader.Read()
	cmn.CheckError(err)

	for {

		tx, err := reader.Read()
		if err == io.EOF {
			fmt.Println("End of stream...")
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

		edges_input <- edge
		// Log - register the event in the sink
		// TODO: Change, the log of events is done here directly on the sink
		log_ch <- edge

		// TODO: Sleep time for debugging to slow down the flux of transactions
		// Leave without this sleep / change it
		time.Sleep(1 * time.Second)

	}

}
