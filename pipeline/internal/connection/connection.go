package connection

/*
Connection module:

Connection management with the static graph database (gdb).
*/

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver neo4j.DriverWithContext
)

// Connection (safe, with godotenv)
// The connection info is not publicly available on the code text file, instead
// it is in user .env file
// Connection through the creation of a DriverWithContext (it allows connections
// and creation of sessions) sessions can be created from it and are cheap
// DriverWithContext objects are immutable, thread-safe, and fairly expensive to
// create, so your application should only create one instance
func SafeConnect() {
	// root context: it has no deadline and it can not be cancelled.
	// used as the base context for connecting to the Neo4j database
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	dbUri := os.Getenv("NEO4J_URI")
	dbUser := os.Getenv("NEO4J_USERNAME")
	dbPassword := os.Getenv("NEO4J_PASSWORD")
	driver, err = neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		panic(err)
	}

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection established.")
}

func WriteQuery(session neo4j.SessionWithContext, query string, params map[string]interface{}) error {

	_, err := neo4j.ExecuteWrite(ctx, session,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, query, params)
			if err != nil {
				return nil, err
			}
			return result, nil
		})

	if err != nil {
		return err
	}

	return nil
}

// TODO: Function to execute a read query -> with ExecuteRead()
// NOTE: The difference between the two is performance reasons:
// https://neo4j.com/docs/go-manual/current/transactions/
/*
Although executing a write query in read mode likely results in a runtime error,
you should not rely on this for access control. The difference between the two
modes is that read transactions will be routed to any node of a cluster, whereas
write ones will be directed to the leader. In other words, there is no guarantee
that a write query submitted in read mode will be rejected.

Similar remarks hold for the .ExecuteRead() and .ExecuteWrite() methods.
*/

func ReadQuery(session neo4j.SessionWithContext,
	query string,
	params map[string]any,
	// function to process result within ReadQuery(), this needs to be done like this since
	// it is not possible to retrieve the result once the transaction is done (outside of
	// the ReadQuery function)
	processResult func(neo4j.ResultWithContext) (any, error)) (any, error) {

	result, err := neo4j.ExecuteRead(ctx, session,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, query, params)
			if err != nil {
				return nil, err
			}
			return processResult(result) // process the result within the active transaction
		})

	return result, err
}

func TestQuery() {
	//session := connection.CreateSession()
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)
	// TODO: Use Indexes for Performance
	// Ensure that the ATM_id field is indexed if you are performing many lookups based on this property.
	// While this is not a different query form, indexing helps improve the performance of queries that filter on this property.
	getATMLocationQuery := `MATCH (a:ATM) WHERE a.ATM_id = $ATM_id RETURN a.loc_latitude AS loc_latitude`

	params := map[string]any{
		"ATM_id": "OGUN-3",
	}

	processCoordinates := func(result neo4j.ResultWithContext) (any, error) {

		//var location float64
		for result.Next(ctx) {
			record := result.Record()

			loc_latitude, found := record.Get("loc_latitude")
			if found {
				fmt.Println("Latitude: ", loc_latitude)
			}

			// location = loc_latitude
		}

		// Check for errors after processing the results
		if err := result.Err(); err != nil {
			return nil, err
		}
		return "done", nil
	}

	result, err := ReadQuery(session, getATMLocationQuery, params, processCoordinates)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Result: ", result)

}

// DriverWithContext VS Sessions
// DriverWithContext:
// - immutable, thread-safe, and fairly expensive to
// create, so your application should only create one instance
// Sessions:
// - created with the method DriverWithContext.NewSession()
// Session creation is a lightweight operation, so sessions can be created and
// destroyed without significant cost. Always close sessions when you are done
// with them.
// not thread safe: you can share the main DriverWithContext object
// across threads, but make sure each routine creates its own sessions.

func CreateSession() neo4j.SessionWithContext {
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	fmt.Println("Session created.")
	return session
}

func CloseSession(session neo4j.SessionWithContext) {
	session.Close(ctx)
	fmt.Println("Session closed.")
}

func CloseConnection() {
	driver.Close(ctx)
	fmt.Println("Connection closed.")
}
