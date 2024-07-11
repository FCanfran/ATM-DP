package dp

import (
	cmn "pipeline/internal/common"
	"bufio"
	"fmt"
	"os"
	"io"
	"encoding/csv"
	"strconv"
	//"strings"
	"time"
)

const channelSize = 5000

// TODO: Set correct time threshold
const timeFilterThreshold = 20

// TODO: Set different times thresholds for the destruction of a filter and
// for the diff transaction times?
const timeTransactionThreshold = 5

type in_comm struct {
	// read-only channels
	Edge          <-chan cmn.Edge // Edges channel
	Front_Channel <-chan in_comm  // in_comm channel
}

func generator(in <-chan cmn.Edge) {
	fmt.Println("...generator creation")
	// Generate input channels
	alerts := make(chan cmn.Graph, channelSize)
	// Note: by default channels created by "make" create bidirectional
	// channels by default. To make it receive-only channel, we perform
	// the type conversion as shown
	tmp := make(chan in_comm, channelSize)
	var front_channels <-chan in_comm // read only channel
	front_channels = tmp

	for {
		select {
		case edge := <-in:
			fmt.Println("generator - edge ", edge.Tx_id, " arrived: ", edge.Number_id, " ", edge.ATM_id)
			fmt.Println("Tx_start:", edge.Tx_start)
			fmt.Println("Tx_end:", edge.Tx_end)
			fmt.Println("Tx_amount:", edge.Tx_amount)
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
			go filter(edge, in, front_channels, new_edge_ch, alerts, new_front_ch)
			// set the new input channels of the generator
			in = new_edge_ch
			front_channels = new_front_ch
		case alert := <-alerts:
			fmt.Println("generator - alert!: Graph", alert)
		case input := <-front_channels:
			// Reconnection of the pipeline (case of a filter having died)
			fmt.Println("generator - reconnection")
			in = input.Edge
			front_channels = input.Front_Channel
		}
	}
}

func filter(edge cmn.Edge, in_edge <-chan cmn.Edge, in_front <-chan in_comm,
	out_edge chan<- cmn.Edge, out_alert chan<- cmn.Graph, out_front chan<- in_comm) {
	
	
	// filter id: is the Card identifier
	var id string = edge.Number_id

	fmt.Println("...filter ", id, "creation - edge arrived: ", edge.Number_id, " ", edge.ATM_id)
	/*
	int_edge := make(chan cmn.Edge, channelSize)
	int_time := make(chan int64) // synchronous
	// TOCHECK: Avoid this channel being blocking (?) Does it make sense?
	int_stop := make(chan bool) // synchronous
	go filter_worker(edge, int_edge, int_time, int_stop, out_alert)

	for {
		select {
		case edge := <-in_edge:
			fmt.Println("filter ", id, " - edge arrived: ", edge.Card, " ", edge.ATM, " ", edge.Time)
			if edge.Card == id {
				fmt.Println("filter ", id, " - same card edge arrived")
				int_edge <- edge
			} else {
				fmt.Println("filter ", id, " - diff card edge arrived")
				out_edge <- edge
				int_time <- edge.Time
				// TODO: avoid this signal (stop) being synchronous! -
				// allow worker to tell at any moment to stop! instead of blocking
				if stop := <-int_stop; stop {
					out_front <- in_comm{in_edge, in_front}
					fmt.Println("filter ", id, " - stop")
					return // finish filter
				}
			}
		case input := <-in_front:
			fmt.Println("filter ", id, " - reconnection")
			// a previous filter died, reconnect pipeline
			in_edge = input.Edge
			in_front = input.Front_Channel
		}
	}
	*/
}
/*
func filter_worker(initial_edge cmn.Edge, int_edge <-chan cmn.Edge, int_time <-chan int64, int_stop chan<- bool,
	out_alert chan<- cmn.Graph) {

	var time int64 = initial_edge.Time
	var edge cmn.Edge = initial_edge
	fmt.Println("...filter_worker creation - edge arrived: ", edge.Card, " ", edge.ATM, " ", edge.Time)
	// TODO: Save more edges? Not only the last one? (the last transaction)
	// var edges []cmn.Edge

	// TODO: this goroutine dies alone after its father (the filter) dies?
	// -> it is the only process with which it has communication / is connected
	for {
		select {
		case new_edge := <-int_edge:
			// TODO: Check for the pattern and output alert in that case
			// --> to develop more...
			// TODO: keep a list of all the edges, not only the last one?
			if (new_edge.ATM != edge.ATM) && (new_edge.Time-edge.Time < timeTransactionThreshold) {
				// alert is the pattern: edge list that form the pattern
				var pattern cmn.Graph // = []Edge
				pattern = append(pattern, edge, new_edge)
				out_alert <- pattern
				// TODO: check timeout in this case? -> no, in general,
				// if we register movement with the card then no need to check
				// timeouts to potentially erase the filter (the filter is active)
			} else {
				// TODO:
				// keep a list of all... if alert do not save the new_edge, otherwise
				// yes... --> think about this
				// For the moment, if alert, then the original edge is saved (the
				// last to arrive is considered erroneous)
				// If not alert, then the last is correct, so we need to update it
				edge = new_edge
				// Update time (note that no check if filter has to die, since
				// it has just been updated by a new transaction on same card - it remains active)
				time = new_edge.Time
				fmt.Println("filter-worker ", " - no alert - update card edge to: ", edge.Card, " ", edge.ATM, " ", edge.Time)
			}

		case new_time := <-int_time:
			// Timeout check: test if the filter has to die, in that case
			// send stop signal to the (father) filter
			if new_time-time > timeFilterThreshold {
				int_stop <- true
			} else {
				int_stop <- false
			}
		}

	}
}
*/

/*
func test_generator(in <-chan cmn.Edge) {
	for edge := range in {
		fmt.Println("c: ", edge.Card, "t: ", edge.Time, "a: ", edge.ATM)
	}

}
*/

func Start(istream string) {

	// Creation of edges channel to pass from the read input to the pipeline
	edges_input := make(chan cmn.Edge, channelSize)

	go generator(edges_input)

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

	// READING ONLY 1 TX
	// -------------------------------------------------------------------------
	//for {
		
		tx, err := reader.Read()
		if err == io.EOF {
			fmt.Println("End of stream...")
			//break
		} 
		cmn.CheckError(err)
		fmt.Println(tx)
			
		// conversions
		tx_id, err := strconv.ParseInt(tx[0], 10, 64) // 10: base (decimal) & 64: bit-size (int64) 
		cmn.CheckError(err)
		// https://yourbasic.org/golang/format-parse-string-time-date-example/
		const layout = "2006-01-02 15:04:05"
		tx_start, err := time.Parse(layout, tx[3])
		cmn.CheckError(err)
		tx_end, err := time.Parse(layout, tx[4])
		cmn.CheckError(err)

		tx_amount, err := strconv.ParseFloat(tx[5], 32)
		cmn.CheckError(err)
		tx_amount_32 := float32(tx_amount)
		
		/*
		type Edge struct {
			Number_id string 	// Card id
			ATM_id string		// ATM id
			Tx_id int64			// transaction id
			Tx_start time.Time	// transaction start date time (DD/MM/YYYY HH:MM:SS)
			Tx_end time.Time	// transaction end date time (DD/MM/YYYY HH:MM:SS)
			Tx_amount float32	// transaction amount 
		}
		*/
		edge := cmn.Edge{
			Number_id: tx[1], 
			ATM_id: tx[2], 
			Tx_id: tx_id,
			Tx_start: tx_start,
			Tx_end: tx_end,
			Tx_amount: tx_amount_32,
		}

		fmt.Println(edge)

		//edges_input <- edge

		//time.Sleep(1 * time.Second)
		
	//}
	// -------------------------------------------------------------------------



}
