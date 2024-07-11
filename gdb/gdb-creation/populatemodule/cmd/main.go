package main

import "populatemodule"

func main() {
	populatemodule.SafeConnect()
	populatemodule.UniquenessConstraints()
	populatemodule.Populate()
	populatemodule.CloseConnection()
}
