/*
Entry point of the program
*/

package main

import (
	"fmt"
	"os"
	"time"

	cmn "pipeline/internal/common"

	"github.com/apache/arrow/go/v11/arrow"
	"github.com/apache/arrow/go/v11/arrow/array"
	"github.com/apache/arrow/go/v11/arrow/csv"
)

// cmn "pipeline/internal/common"
// 	"encoding/csv"
//"github.com/apache/arrow/go/arrow/array"
//"github.com/apache/arrow/go/arrow/csv"

var chunkSize int = 5

/*
// Opt A: arrow/csv - sending records from worker to main
// TOCHECK: Not working - investigate more -> not all the records are arriving to the main
// process
func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <executionDescriptionFile>")
		return
	}

	// test - using:
	// - arrow/csv for file input reading by chunks of csv rows
	// - encoding/csv to write in an output file
	// _____________
	// worker: puts the csv row chunks in memory - giving them to the main process
	// main: receives the csv rows, process them

	// channel of chunks
	chunk_ch := make(chan array.Record)

	// TODO: Pass an array.record = chunk of rows through the channel
	// --> later in the main, extract each row as a []string and convert
	// correspondingly with the same process as before

	go func() {

		file, err := os.Open(os.Args[1])
		cmn.CheckError(err)
		defer file.Close()
		cmn.CheckError(err)

		// opc 1: read all as strings - later we will do the conversions
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "transaction_id", Type: arrow.BinaryTypes.String},
				{Name: "number_id", Type: arrow.BinaryTypes.String},
				{Name: "ATM_id", Type: arrow.BinaryTypes.String},
				{Name: "transaction_type", Type: arrow.BinaryTypes.String},
				{Name: "transaction_start", Type: arrow.BinaryTypes.String},
				{Name: "transaction_end", Type: arrow.BinaryTypes.String},
				{Name: "transaction_amount", Type: arrow.BinaryTypes.String},
			},
			nil,
		)

		r := csv.NewReader(file, schema, csv.WithHeader(true), csv.WithChunk(4))
		defer r.Release()

		n := 0
		indrec := 0
		var recs [2]array.Record
		for r.Next() {
			recs[indrec] = r.Record()
			for i, col := range recs[indrec].Columns() {
				fmt.Printf("rec[%d][%q]: %v\n", n, recs[indrec].ColumnName(i), col)
			}
			chunk_ch <- recs[indrec]
			n++
			indrec = (indrec + 1) % 2
		}

		// check for reader errors indicating issues converting csv values
		// to the arrow schema types
		err = r.Err()
		if err != nil {
			log.Fatal(err)
		}
	}()

	i := 0

	for {
		chunk, ok := <-chunk_ch
		if !ok {
			fmt.Println("Stream - !ok in chunk_ch channel")
		}

		fmt.Println("+++++++++++++++++ chunk i: ", i, " +++++++++++++++++++++")

		for i, col := range chunk.Columns() {
			fmt.Printf("rec[%d][%q]: %v\n", 9, chunk.ColumnName(i), col)
		}
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++")
		i++
	}

}
*/

// Opt A.1: apache/arrow -> reading chunks of csv rows in the worker and giving them to the main
// - Reading as corresponding data types directly
// - Transposing back to rows (as the library optimizes saving the csv by columns)
func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <executionDescriptionFile>")
		return
	}
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
				{Name: "transaction_start", Type: &arrow.TimestampType{Unit: arrow.Second, TimeZone: "UTC"}},
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
				cmn.PrintEdgeComplete("", event.E)
				edgeEvents = append(edgeEvents, event)
			}

			chunk_ch <- edgeEvents
			edgeEvents = nil // clear the rows holder

		}

		if reader.Err() != nil {
			fmt.Printf("error: %s ", reader.Err().Error())
		}

		close(chunk_ch)
		fmt.Println(";;;;;;;;;;;; worker ends ;;;;;;;;;;;;;;;;")
	}()

	i := 0

	for chunk := range chunk_ch {

		fmt.Println("+++++++++++++++++ chunk i: ", i, " +++++++++++++++++++++")
		for _, event := range chunk {
			cmn.PrintEdgeComplete("", event.E)
		}
		i++
		fmt.Println("........................................................")
	}

	fmt.Println(";;;;;;;;;;;; main ends ;;;;;;;;;;;;;;;;")

}

/*
// Opt A: apache/arrow -> reading chunks of csv rows in the worker and giving them to the main
// - Reading as string types first and afterwards converting to the adequate data types.
// - Transposing back to rows (as the library optimizes saving the csv by columns)
func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <executionDescriptionFile>")
		return
	}

	// worker: puts the csv row chunks in memory
	// - giving them to the main process.
	// - read as strings
	// - transpose to form back the rows...
	// main: receives the csv rows, process them. Do the conversion types here.

	// channel of chunks - slices of rows

	chunk_ch := make(chan [][]string)

	go func() {

		file, err := os.Open(os.Args[1])
		cmn.CheckError(err)
		defer file.Close()
		cmn.CheckError(err)

		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "transaction_id", Type: arrow.BinaryTypes.String},
				{Name: "number_id", Type: arrow.BinaryTypes.String},
				{Name: "ATM_id", Type: arrow.BinaryTypes.String},
				{Name: "transaction_type", Type: arrow.BinaryTypes.String},
				{Name: "transaction_start", Type: arrow.BinaryTypes.String},
				{Name: "transaction_end", Type: arrow.BinaryTypes.String},
				{Name: "transaction_amount", Type: arrow.BinaryTypes.String},
			},
			nil,
		)

		r := csv.NewReader(file, schema, csv.WithHeader(true), csv.WithChunk(chunkSize))
		defer r.Release()

		var rows [][]string
		var rec array.Record
		for r.Next() {

			rec = r.Record()

			// obtain the rows - transposing them back to row form
			numRows := int(rec.NumRows())
			for i := 0; i < numRows; i++ {
				// Extract row values
				row := make([]string, rec.NumCols())
				for j := 0; j < int(rec.NumCols()); j++ {
					// For each column in the record, get the value for the current row index
					row[j] = fmt.Sprintf("%v", rec.Column(j).(*array.String).Value(i))
				}

				// Print the row
				fmt.Printf("Row %d: %v\n", i, row)
				rows = append(rows, row)
			}

			chunk_ch <- rows
			rows = nil // clear the rows holder
		}

		close(chunk_ch)
	}()

	i := 0

	for chunk := range chunk_ch {

		fmt.Println("+++++++++++++++++ chunk i: ", i, " +++++++++++++++++++++")
		for _, row := range chunk {
			event := cmn.ReadEdge(row) // converting to corresp. types and creating edge event
			cmn.PrintEdgeComplete("", event.E)
		}
		i++
	}
}
*/

/*
// Opt B: encoding/csv -> reading chunks of csv rows in the worker and giving them to the main
func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <executionDescriptionFile>")
		return
	}

	// test - using:
	// - arrow/csv for file input reading by chunks of csv rows
	// - encoding/csv to write in an output file
	// _____________
	// worker: puts the csv row chunks in memory - giving them to the main process
	// main: receives the csv rows, process them

	// channel of chunks

	chunk_ch := make(chan [][]string)

	go func() {

		file, err := os.Open(os.Args[1])
		cmn.CheckError(err)
		defer file.Close()
		cmn.CheckError(err)

		// csv reader
		reader := csv.NewReader(bufio.NewReader(file))
		_, err = reader.Read() // Read and discard the header line
		cmn.CheckError(err)

		var rows [][]string
		var chunkSize int = 5
		i := 0
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			cmn.CheckError(err)

			// Print the row
			fmt.Printf("Row %d: %v\n", i, row)
			rows = append(rows, row)
			i++
			if i == chunkSize {
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

	i := 0

	for chunk := range chunk_ch {
		//chunk, ok := <-chunk_ch
		//if !ok {
		//	fmt.Println("Stream - !ok in chunk_ch channel")
		//}

		fmt.Println("+++++++++++++++++ chunk i: ", i, " +++++++++++++++++++++")
		for _, row := range chunk {
			event := cmn.ReadEdge(row) // converting to corresp. types and creating edge event
			cmn.PrintEdgeComplete("", event.E)
		}
		i++
	}
}
*/
