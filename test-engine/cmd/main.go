/*
Entry point of the program
*/

package main

import (
	"fmt"
	"log"
	"os"
	cmn "pipeline/internal/common"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/csv"
)

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
		for r.Next() {
			rec := r.Record()
			fmt.Println("¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬")
			fmt.Println(rec)
			fmt.Println("¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬¬")
			for i, col := range rec.Columns() {
				fmt.Printf("rec[%d][%q]: %v\n", n, rec.ColumnName(i), col)
			}
			n++
		}

		// check for reader errors indicating issues converting csv values
		// to the arrow schema types
		err = r.Err()
		if err != nil {
			log.Fatal(err)
		}
	}()

}
