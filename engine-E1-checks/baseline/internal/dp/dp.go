package dp

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	cmn "pipeline/internal/common"
	"time"
)

func Stream(start_time time.Time, istream string, out_stream chan<- cmn.Event) {

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
			event = cmn.ReadEdge(row)
			t := time.Since(start_time)
			event.Timestamp = t
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
