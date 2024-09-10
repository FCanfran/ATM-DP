/*
Entry point of the program
*/

package main

import (
	"pipeline/internal/connection"
)

func main() {

	// Connection to static gdb
	connection.SafeConnect()

	// --------------- TEST: THIS DOES NOT GO HERE! -------------//

	// TODO: Put this in each of the filters!

	// TODO: Create a session in each of the filters - NOTE THAT:
	// Session creation is a lightweight operation, so sessions can be created and
	// destroyed without significant cost. Always close sessions when you are done
	// with them. However they are not thread safe: you can share the main DriverWithContext object
	// across threads, but make sure each routine creates its own sessions.

	// Ideally:
	// -- DriverWithContext object: only 1 (in the connection module)
	// -- Sessions: 1 per filter

	// TODO: Use Indexes for Performance
	// Ensure that the ATM_id field is indexed if you are performing many lookups based on this property.
	// While this is not a different query form, indexing helps improve the performance of queries that filter on this property.
	session := connection.CreateSession()
	getATMLocationQuery := `MATCH (a:ATM) WHERE a.ATM_id = $ATM_id RETURN a.loc_latitude, a.loc_longitude`
	params := map[string]interface{}{
		"ATM_id": "OGUN-3",
	}
	connection.ReadQuery(session, getATMLocationQuery, params)
	/*
		if result.Next(connection.Ctx) {
			fmt.Println("@@@@@@@@@@@@")
			record := result.Record()
			loc_latitude, _ := record.Get("a.loc_latitude")
			fmt.Println(loc_latitude)
		}
	*/

	connection.CloseSession(session)
	// ----------------------------------------------------------//

	/*
		// obtain stream fileName from args
		istream := os.Args[1]

		fmt.Println("Reading stream from: ", istream)

		// start the pipeline and give the stream edge by edge
		dp.Start(istream)
	*/
}
