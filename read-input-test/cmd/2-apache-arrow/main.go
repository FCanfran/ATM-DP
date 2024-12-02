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
)

// Opt A.2 apache/arrow -> reading chunks of csv rows in the worker and giving them to the main
// - Reading as string types first and afterwards converting to the adequate data types.
// - Transposing back to rows (as the library optimizes saving the csv by columns)
func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go csvFile chunkSize")
		return
	}

	var chunkSize int
	chunkSize, err := strconv.Atoi(os.Args[2])
	cmn.CheckError(err)

	start := time.Now()
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

		reader := csv.NewReader(file, schema, csv.WithHeader(true), csv.WithChunk(chunkSize))
		defer reader.Release()

		var rows [][]string
		var rec arrow.Record
		for reader.Next() {

			rec = reader.Record()

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

		if reader.Err() != nil {
			fmt.Printf("error: %s ", reader.Err().Error())
		}

		close(chunk_ch)
	}()

	i := 0
	rows := 0

	for chunk := range chunk_ch {
		fmt.Println("+++++++++++++++++ chunk i: ", i, " +++++++++++++++++++++")
		for _, row := range chunk {
			event := cmn.ReadEdge(row) // converting to corresp. types and creating edge event
			cmn.PrintEdgeComplete("", event.E)
			rows++
		}
		i++
	}

	t := time.Since(start)
	fmt.Println("Total num of rows read: ", rows)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")
}
