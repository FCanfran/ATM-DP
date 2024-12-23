package connection

/*
Connection module:

Connection management with the static graph database (gdb).
*/

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver neo4j.DriverWithContext
	// TODO: Leave as global variable or not
	//ctx context.Context
)

// Connection (safe, with godotenv)
// The connection info is not publicly available on the code text file, instead
// it is in user .env file
// Connection through the creation of a DriverWithContext (it allows connections
// and creation of sessions) sessions can be created from it and are cheap
// DriverWithContext objects are immutable, thread-safe, and fairly expensive to
// create, so your application should only create one instance
func SafeConnect() context.Context {
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

	// Validate that the variables are set
	if dbUri == "" || dbUser == "" || dbPassword == "" {
		log.Fatal("Missing required environment variables: NEO4J_URI, NEO4J_USERNAME, or NEO4J_PASSWORD - use a .env file to specify them")
	}

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

	return ctx
}

func WriteQuery(ctx context.Context,
	session neo4j.SessionWithContext,
	query string,
	params map[string]any) error {

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

func ReadQuery(ctx context.Context,
	session neo4j.SessionWithContext,
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

func CreateSession(ctx context.Context) neo4j.SessionWithContext {
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	fmt.Println("Session created.")
	return session
}

func CloseSession(ctx context.Context, session neo4j.SessionWithContext) {
	session.Close(ctx)
	fmt.Println("Session closed.")
}

// TODO/TOCHECK: ctx as a global variable or not

func CloseConnection(ctx context.Context) {
	driver.Close(ctx)
	fmt.Println("Connection closed.")
}

/*
func CloseConnection() {
	driver.Close(ctx)
	fmt.Println("Connection closed.")
}
*/
