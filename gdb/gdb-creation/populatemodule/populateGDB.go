package populatemodule

import (
	"context"
	"fmt"
	"os"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver neo4j.DriverWithContext
	ctx    context.Context
)

/*
// Connection
func Connect() {

	//- Context package provides a way to pass cancellation signals and deadlines
	// across API boundaries to control the lifetime of operations and manage
	// resources efficiently.
	// - A context represents the context in which a particular operation is being
	// executed. It allows you to control the execution of various operations, such
	// as HTTP requests, database queries, or any long-running tasks, in a way that
	// allows for cancellation, timeouts, and passing of additional information.

	ctx = context.Background()
	// URI examples: "neo4j://localhost", "neo4j+s://xxx.databases.neo4j.io"
	dbUri := "neo4j+s://e2c3d8d5.databases.neo4j.io"
	dbUser := "neo4j"
	dbPassword := "HSJzzC4ySP8xG_OdWYwOn9cFs_gBVPh2EYbm3f1SgTU"

	// Creating a DriverWithContext instance only provides information on how to
	// access the database, but does not actually establish a connection.
	// Connection is instead deferred to when the first query is executed.

	driver, err = neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	defer driver.Close(ctx)

	// Always close DriverWithContext objects to free up all allocated resources,
	// even upon unsuccessful connection or runtime errors in subsequent querying.
	// The safest practice is to defer the call to DriverWithContext.Close(ctx)
	// after the object is successfully created.

	// Verify immediately that the driver can connect to the database
	// -> after initializing the driver
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection established.")
}
*/

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

func writeQuery(session neo4j.SessionWithContext, query string) error {

	_, err := neo4j.ExecuteWrite(ctx, session,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, query, nil)
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

func populateATMs(session neo4j.SessionWithContext) {
	query := `
	LOAD CSV WITH HEADERS FROM 'file:///csv/atm.csv' AS row
	MERGE (a:ATM {
		ATM_id: row.ATM_id,
		loc_latitude: toFloat(row.loc_latitude),
		loc_longitude: toFloat(row.loc_longitude),
		city: row.city,
		country: row.country
	});
	`
	err := writeQuery(session, query)
	if err != nil {
		fmt.Println("ATM population: failure - %v", err)
	} else {
		fmt.Println("ATM population: sucessful")
	}
}

func populateBanks(session neo4j.SessionWithContext) {
	query := `
	LOAD CSV WITH HEADERS FROM 'file:///csv/bank.csv' AS row
	MERGE (b:Bank {
		name: row.name, 
		code: row.code, 
		loc_latitude: toFloat(row.loc_latitude), 
		loc_longitude: toFloat(row.loc_longitude)
	});
	`
	err := writeQuery(session, query)
	if err != nil {
		fmt.Println("Bank population: failure - %v", err)
	} else {
		fmt.Println("Bank population: sucessful")
	}
}

func populateATMBanks(session neo4j.SessionWithContext) {
	query := `
	LOAD CSV WITH HEADERS FROM 'file:///csv/atm-bank.csv' AS row
             MATCH (a:ATM {ATM_id: row.ATM_id})
             MATCH (b:Bank {code: row.code})
             MERGE (a)-[r:BELONGS_TO]->(b);
	`
	err := writeQuery(session, query)
	if err != nil {
		fmt.Println("ATM-Bank relationships population: failure - %v", err)
	} else {
		fmt.Println("ATM-Bank relationships population: sucessful")
	}
}

func populateCards(session neo4j.SessionWithContext) {
	query := `
	LOAD CSV WITH HEADERS FROM 'file:///csv/card.csv' AS row
	MERGE (c:Card {
		number_id: row.number_id, 
		client_id: row.client_id, 
		expiration: date(row.expiration), 
		CVC: toInteger(row.CVC), 
		extract_limit: toFloat(row.extract_limit), 
		loc_latitude: toFloat(row.loc_latitude), 
		loc_longitude: toFloat(row.loc_longitude)});
	`
	err := writeQuery(session, query)
	if err != nil {
		fmt.Println("Card population: failure - %v", err)
	} else {
		fmt.Println("Card population: sucessful")
	}
}

func populateCardBanks(session neo4j.SessionWithContext) {
	query := `
	LOAD CSV WITH HEADERS FROM 'file:///csv/card-bank.csv' AS row
             MATCH (c:Card {number_id: row.number_id})
             MATCH (b:Bank {code: row.code})
             MERGE (c)-[r:ISSUED_BY]->(b);
	`
	err := writeQuery(session, query)
	if err != nil {
		fmt.Println("Card-Bank relationships population: failure - %v", err)
	} else {
		fmt.Println("Card-Bank relationships population: sucessful")
	}
}

func Populate() {
	fmt.Println("Population of the GDB...")
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	populateATMs(session)
	populateBanks(session)
	populateATMBanks(session)
	populateCards(session)
	populateCardBanks(session)

}

// Creates uniqueness constraints within each kind of node's IDs
// Bank, ATM and Card
// Avoids duplication of nodes prior to the population and also afterwards, whenever for
// example an ATM node with the same ATM_id as an already existing one in the database
// wants to be inserted. This is forbidden!.
func UniquenessConstraints() {
	fmt.Println("Adding uniqueness constraints on the nodes IDs to the GDB...")
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	bankUniquenessQuery := `
		CREATE CONSTRAINT code IF NOT EXISTS
		FOR (b:Bank) REQUIRE b.code IS UNIQUE;
	`
	err := writeQuery(session, bankUniquenessQuery)
	if err != nil {
		fmt.Println("Bank uniqueness constraint addition: failure - %v", err)
	} else {
		fmt.Println("Bank uniqueness constraint addition: sucess")
	}

	cardUniquenessQuery := `
		CREATE CONSTRAINT number_id IF NOT EXISTS
		FOR (c:Card) REQUIRE c.number_id IS UNIQUE;
	`
	err = writeQuery(session, cardUniquenessQuery)
	if err != nil {
		fmt.Println("Card uniqueness constraint addition: failure - %v", err)
	} else {
		fmt.Println("Card uniqueness constraint addition: sucess")
	}


	ATMUniquenessQuery := `
		CREATE CONSTRAINT ATM_id IF NOT EXISTS
		FOR (a:ATM) REQUIRE a.ATM_id IS UNIQUE;
	`
	err = writeQuery(session, ATMUniquenessQuery)
	if err != nil {
		fmt.Println("ATM uniqueness constraint addition: failure - %v", err)
	} else {
		fmt.Println("ATM uniqueness constraint addition: sucess")
	}
}

// Query
// Once connected, run queries using Cypher and the function ExecuteQuery()
// - Read (MATCH), write (MERGE), update (MATCH + SET), delete (DETACH DELETE)
func Query() {

}

// Run transaction
// A transaction is a unit of work that is either committed in its entirety or
// rolled back on failure. Use the methods SessionWithContext.ExecuteRead() and
// SessionWithContext.ExecuteWrite() to run managed transactions.
// Sessions are created with the method DriverWithContext.NewSession()
// Session creation is a lightweight operation, so sessions can be created and
// destroyed without significant cost. Always close sessions when you are done
// with them.
// Sessions are not thread safe: you can share the main DriverWithContext object
// across threads, but make sure each routine creates its own sessions.
// DIFFERENCE WITH EXECUTEQUERY --> think of managed transactions as a way of
// unwrapping the flow of ExecuteQuery() and being able to specify its desired behavior 
// in more places.
func Transaction() {
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)
	title, _ := session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, `
			MATCH (p:Person {name:'Hugo Weaving'}) -[:ACTED_IN]->(m:Movie) RETURN m.title AS title
				`, map[string]any{
				"filter": "Al",
			})
			records, _ := result.Collect(ctx)
			return records, nil
		})
	for _, title := range title.([]*neo4j.Record) {
		fmt.Println(title.AsMap())
	}
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

func CloseConnection() {
	driver.Close(ctx)
	fmt.Println("Connection closed.")
}
