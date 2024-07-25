package main

import "populatemodule/internal/populate"

func main() {
	populate.SafeConnect()
	populate.UniquenessConstraints()
	populate.Populate()
	populate.CloseConnection()
}
