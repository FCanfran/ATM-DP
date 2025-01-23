package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go numRows")
		return
	}

	numRows, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error: invalid number of rows")
		return
	}

	outputDir := "output"

	_, err = os.Stat(outputDir)

	// If the directory does not exist, err will be nil and we can create it
	if os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}

	outFileName := outputDir + "/csv" + os.Args[1] + ".csv"
	file, err := os.Create(outFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"transaction_id", "number_id", "ATM_id", "transaction_type", "transaction_start", "transaction_end", "transaction_amount",
	}
	writer.Write(header)

	tx_start := []string{
		"31", "c-NIGER-7", "EXT-3", "0", "2018-04-04 03:54:48", "", "",
	}

	tx_end := []string{
		"12", "c-NIGER-2", "NIGER-5", "0", "2018-04-04 01:42:08", "2018-04-04 01:43:39", "33252.26",
	}

	// Replicate the lines x times
	for i := 0; i < numRows/2; i++ {
		// Write the first line
		err := writer.Write(tx_start)
		if err != nil {
			fmt.Println("Error writing tx_start:", err)
			return
		}
		// Write the second line
		err = writer.Write(tx_end)
		if err != nil {
			fmt.Println("Error writing tx_end:", err)
			return
		}
	}

	fmt.Println("Created: ", outFileName)
}
