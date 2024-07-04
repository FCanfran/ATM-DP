package main

import "populatemodule"

func main() {
	populatemodule.SafeConnect()
	populatemodule.Populate()
	populatemodule.CloseConnection()
}
