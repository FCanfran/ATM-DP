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

// Opt B.2: encoding/csv -> no worker
func main() {

	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go csvFile chunkSize outFile")
		return
	}

	var chunkSize int
	chunkSize, err := strconv.Atoi(os.Args[2])
	cmn.CheckError(err)

	fmt.Println(chunkSize)

	start := time.Now()
	// channel of chunks - slices of rows

	file, err := os.Open(os.Args[1])
	cmn.CheckError(err)
	defer file.Close()
	cmn.CheckError(err)

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read() // Read and discard the header line
	cmn.CheckError(err)

	rows := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		cmn.CheckError(err)

		event := cmn.ReadEdge(row) // converting to corresp. types and creating edge event
		_ = event
		rows++
	}

	t := time.Since(start)
	fmt.Println("Total num of rows read: ", rows)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")

	// Write results to csv outputfile
	file, err = os.OpenFile(os.Args[3], os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Error opening/creating file: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	/*
		row := []string{
			strconv.Itoa(chunkSize),
			strconv.FormatInt(t.Milliseconds(), 10),
		}
	*/
	row := []string{
		strconv.Itoa(rows),
		strconv.FormatInt(t.Milliseconds(), 10),
	}

	// Write the row to the CSV
	if err := writer.Write(row); err != nil {
		fmt.Printf("Error writing to CSV: %v\n", err)
		return
	}
}
