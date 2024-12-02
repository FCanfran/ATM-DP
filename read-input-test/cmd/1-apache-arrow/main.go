/*
Entry point of the program
*/

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	cmn "pipeline/internal/common"

	"github.com/apache/arrow/go/v11/arrow"
	"github.com/apache/arrow/go/v11/arrow/array"
	"github.com/apache/arrow/go/v11/arrow/csv"

	encodingcsv "encoding/csv"
)

// cmn "pipeline/internal/common"
// 	"encoding/csv"
//"github.com/apache/arrow/go/arrow/array"
//"github.com/apache/arrow/go/arrow/csv"

// Opt A.1: apache/arrow -> reading chunks of csv rows in the worker and giving them to the main
// - Reading as corresponding data types directly
// - Transposing back to rows (as the library optimizes saving the csv by columns)
func main() {

	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go csvFile chunkSize outFile")
		return
	}

	var chunkSize int
	chunkSize, err := strconv.Atoi(os.Args[2])
	cmn.CheckError(err)

	start := time.Now()
	// channel of chunks - slices of edge events
	chunk_ch := make(chan []cmn.Event)

	go func() {

		file, err := os.Open(os.Args[1])
		cmn.CheckError(err)
		defer file.Close()
		cmn.CheckError(err)

		// types: https://github.com/apache/arrow/blob/bc219186db40/go/arrow/datatype_numeric.gen.go#L37
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "transaction_id", Type: arrow.PrimitiveTypes.Int32},
				{Name: "number_id", Type: arrow.BinaryTypes.String},
				{Name: "ATM_id", Type: arrow.BinaryTypes.String},
				{Name: "transaction_type", Type: arrow.PrimitiveTypes.Uint8},
				{Name: "transaction_start", Type: arrow.FixedWidthTypes.Timestamp_s},
				{Name: "transaction_end", Type: arrow.FixedWidthTypes.Timestamp_s},
				{Name: "transaction_amount", Type: arrow.PrimitiveTypes.Float32},
			},
			nil,
		)

		reader := csv.NewReader(file, schema, csv.WithHeader(true), csv.WithChunk(chunkSize), csv.WithNullReader(true, ""))
		defer reader.Release()

		var edgeEvents []cmn.Event
		var rec arrow.Record

		for reader.Next() {

			rec = reader.Record()

			// obtain the rows - transposing them back to row edge form
			numRows := int(rec.NumRows())
			for i := 0; i < numRows; i++ {

				tx_id := rec.Column(0).(*array.Int32).Value(i)
				number_id := rec.Column(1).(*array.String).Value(i)
				ATM_id := rec.Column(2).(*array.String).Value(i)
				tx_type := cmn.TxType(rec.Column(3).(*array.Uint8).Value(i))

				var tx_start time.Time
				tx_start_col := rec.Column(4).(*array.Timestamp)
				if tx_start_col.IsNull(i) {
					tx_start = time.Time{} // Zero value for time.Time
				} else {
					tx_start_seconds := tx_start_col.Value(i)
					tx_start = time.Unix(int64(tx_start_seconds), 0).UTC()
				}

				var tx_end time.Time
				tx_end_col := rec.Column(5).(*array.Timestamp)
				if tx_end_col.IsNull(i) {
					tx_end = time.Time{} // Zero value for time.Time
				} else {
					tx_end_seconds := tx_end_col.Value(i)
					tx_end = time.Unix(int64(tx_end_seconds), 0).UTC()
				}

				tx_amount := rec.Column(6).(*array.Float32).Value(i)

				edge := cmn.Edge{
					Number_id: number_id,
					ATM_id:    ATM_id,
					Tx_id:     tx_id,
					Tx_type:   tx_type,
					Tx_start:  tx_start,
					Tx_end:    tx_end,
					Tx_amount: tx_amount,
				}

				// check if tx_end is zero-value time (null)
				var eventType cmn.EventType
				if tx_end.Equal(time.Time{}) {
					eventType = cmn.EdgeStart
				} else {
					eventType = cmn.EdgeEnd
				}

				event := cmn.Event{
					Type: eventType,
					E:    edge,
				}

				// Print the row
				//cmn.PrintEdgeComplete("", event.E)
				edgeEvents = append(edgeEvents, event)
			}

			chunk_ch <- edgeEvents
			edgeEvents = nil // clear the rows holder

		}

		if reader.Err() != nil {
			fmt.Printf("error: %s ", reader.Err().Error())
		}

		close(chunk_ch)
	}()

	i := 0
	rows := 0

	for chunk := range chunk_ch {
		//fmt.Println("+++++++++++++++++ chunk i: ", i, " +++++++++++++++++++++")
		for _, event := range chunk {
			//cmn.PrintEdgeComplete("", event.E)
			_ = event
			rows++
		}
		i++
	}

	t := time.Since(start)
	fmt.Println("Total num of rows read: ", rows)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")

	// Write results to csv outputfile
	file, err := os.OpenFile(os.Args[3], os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Error opening/creating file: %v\n", err)
		return
	}
	defer file.Close()

	writer := encodingcsv.NewWriter(file)
	defer writer.Flush()

	// Data to write
	row := []string{
		strconv.Itoa(chunkSize),
		strconv.FormatInt(t.Milliseconds(), 10),
	}

	// Write the row to the CSV
	if err := writer.Write(row); err != nil {
		fmt.Printf("Error writing to CSV: %v\n", err)
		return
	}

}
