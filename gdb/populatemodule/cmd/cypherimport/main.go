package main

import (
	"fmt"
	"os"
	"populatemodule/internal/populate"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <csvPath>")
		return
	}

	csvPath := os.Args[1]

	populate.SafeConnect()
	populate.UniquenessConstraints()
	populate.PopulateAlt(csvPath)
	populate.CloseConnection()
}
