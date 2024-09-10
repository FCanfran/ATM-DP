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
	// TODO: Como apaño lo he puesto público para facilitar acceso desde el resto de módulos
	// --> Estructurar y modular de la forma correcta sin jugarretas!!!!
	// Define a module that holds and manages the context and this variables (this one for
	// example) and define Init() and other proper functions to correctly manage them
	//Ctx context.Context
	ctx context.Context
)

// Connection (safe, with godotenv)
// The connection info is not publicly available on the code text file, instead
// it is in user .env file
// Connection through the creation of a DriverWithContext (it allows connections
// and creation of sessions) sessions can be created from it and are cheap
// DriverWithContext objects are immutable, thread-safe, and fairly expensive to
// create, so your application should only create one instance
func SafeConnect() {
	ctx = context.Background()
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

// TODO: Devolver el resultado!!
func ReadQuery(session neo4j.SessionWithContext, query string, params map[string]interface{}) (neo4j.ResultWithContext, error) {

	res, err := neo4j.ExecuteRead(ctx, session,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, query, params)
			if err != nil {
				return nil, err
			}
			// TODO: Process the result:
			// How to do it?: https://neo4j.com/docs/go-manual/current/transactions/
			// example on the "run a managed transaction" section
			// Collect() retrieves all records into a list - SEE OTHER ALTERNATIVES!?
			// records, err := result.Collect(ctx) --> no need to: many disadvantages:
			// - Memory usage: if result set is large, collecting all records into memory might lead to high
			//   memory consumption and potential performance issues.
			// - Performance Overhead: Collecting all results first can be less efficient than processing
			// records one by one, especially if you only need a small subset of results.
			return result, nil
		})

	if err != nil {
		return nil, err
	}
	// Assert the type of result
	result, ok := res.(neo4j.ResultWithContext)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	if result.Next(ctx) {
		record := result.Record()
		loc_latitude, _ := record.Get("a.loc_latitude")
		fmt.Println(loc_latitude)
	}

	fmt.Println(result.Next(ctx))

	return result, nil
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
