/*
Entry point of the program
*/

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	cmn "pipeline/internal/common"
)

// Opt B: encoding/csv -> reading chunks of csv rows in the worker and giving them to the main
// - arrow/csv for file input reading by chunks of csv rows
// - encoding/csv to write in an output file
// _____________
// worker: puts the csv row chunks in memory - giving them to the main process
// main: receives the csv rows, process them
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
	rows := 0

	for chunk := range chunk_ch {
		fmt.Println("+++++++++++++++++ chunk i: ", i, " +++++++++++++++++++++")
		for _, row := range chunk {
			event := cmn.ReadEdge(row) // converting to corresp. types and creating edge event
			//cmn.PrintEdgeComplete("", event.E)
			_ = event
			rows++
		}
		i++
	}

	t := time.Since(start)
	fmt.Println("Total num of rows read: ", rows)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")
}
