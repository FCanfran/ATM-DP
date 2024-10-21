package populate

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

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
	dbPassword := ""

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

func writeQuery(session neo4j.SessionWithContext, query string, params map[string]interface{}) error {

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
	err := writeQuery(session, query, nil) // no params (nil)
	if err != nil {
		fmt.Printf("ATM population: failure - %v\n", err)
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
	err := writeQuery(session, query, nil)
	if err != nil {
		fmt.Printf("Bank population: failure - %v\n", err)
	} else {
		fmt.Println("Bank population: sucessful")
	}
}

func populateATMBanksInternal(session neo4j.SessionWithContext) {
	query := `
	LOAD CSV WITH HEADERS FROM 'file:///csv/atm-bank-internal.csv' AS row
             MATCH (a:ATM {ATM_id: row.ATM_id})
             MATCH (b:Bank {code: row.code})
             MERGE (a)-[r:BELONGS_TO]->(b);
	`
	err := writeQuery(session, query, nil)
	if err != nil {
		fmt.Printf("ATM-Bank-Internal relationships population: failure - %v\n", err)
	} else {
		fmt.Println("ATM-Bank-Internal relationships population: sucessful")
	}
}

func populateATMBanksExternal(session neo4j.SessionWithContext) {
	query := `
	LOAD CSV WITH HEADERS FROM 'file:///csv/atm-bank-external.csv' AS row
             MATCH (a:ATM {ATM_id: row.ATM_id})
             MATCH (b:Bank {code: row.code})
             MERGE (a)-[r:EXTERNAL]->(b);
	`
	err := writeQuery(session, query, nil)
	if err != nil {
		fmt.Printf("ATM-Bank-External relationships population: failure - %v\n", err)
	} else {
		fmt.Println("ATM-Bank-External relationships population: sucessful")
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
	err := writeQuery(session, query, nil)
	if err != nil {
		fmt.Printf("Card population: failure - %v\n", err)
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
	err := writeQuery(session, query, nil)
	if err != nil {
		fmt.Printf("Card-Bank relationships population: failure - %v\n", err)
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
	populateATMBanksInternal(session)
	populateATMBanksExternal(session)
	populateCards(session)
	populateCardBanks(session)

}

// ------------------------------------------------------------------------------------------- //
// Alternative CSV population:
// - reading directly the csv with golang and then adding the data directly in neo4j using
//   cypher commands

func populateATMsAlt(session neo4j.SessionWithContext, csvPath string) {
	// 1. Open and read the CSV file
	file, err := os.Open(csvPath + "/atm.csv")
	if err != nil {
		fmt.Printf("could not open CSV file: %v\n", err)
		return
	}
	// closes the file after read from it no matter if there is error or not
	defer file.Close()

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	// Read and discard the header line
	_, err = reader.Read()
	if err != nil {
		fmt.Println("could not read header from CSV file: %w", err)
		return
	}

	query := `
		MERGE (a:ATM {
			ATM_id: $ATM_id,
			loc_latitude: $loc_latitude,
			loc_longitude: $loc_longitude,
			city: $city,
			country: $country
		});
	`

	i := 0
	success := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		i += 1

		ATM_id := row[0]
		loc_latitude, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			fmt.Printf("invalid latitude value at row %d: %v\n", i+2, err)
			continue
		}
		loc_longitude, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			fmt.Printf("invalid longitude value at row %d: %v\n", i+2, err)
			continue
		}
		city := row[3]
		country := row[4]

		params := map[string]interface{}{
			"ATM_id":        ATM_id,
			"loc_latitude":  loc_latitude,
			"loc_longitude": loc_longitude,
			"city":          city,
			"country":       country,
		}

		err = writeQuery(session, query, params)
		if err != nil {
			fmt.Printf("ATM population: failure - %v\n", err)
		} else {
			success += 1
		}
	}

	fmt.Printf("ATM population: %d sucess / %d total\n", success, i)

}

func populateBanksAlt(session neo4j.SessionWithContext, csvPath string) {

	// 1. Open and read the CSV file
	file, err := os.Open(csvPath + "/bank.csv")
	if err != nil {
		fmt.Printf("could not open CSV file: %v\n", err)
		return
	}
	// closes the file after read from it no matter if there is error or not
	defer file.Close()

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	// Read and discard the header line
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("could not read header from CSV file: %v\n", err)
		return
	}

	query := `
		MERGE (b:Bank {
			name: $name, 
			code: $code, 
			loc_latitude: $loc_latitude, 
			loc_longitude: $loc_longitude
		});
	`

	i := 0
	success := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		i += 1

		name := row[0]
		code := row[1]
		loc_latitude, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			fmt.Printf("invalid latitude value at row %d: %v\n", i+2, err)
			continue
		}
		loc_longitude, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			fmt.Printf("invalid longitude value at row %d: %v\n", i+2, err)
			continue
		}

		params := map[string]interface{}{
			"name":          name,
			"code":          code,
			"loc_latitude":  loc_latitude,
			"loc_longitude": loc_longitude,
		}

		err = writeQuery(session, query, params)
		if err != nil {
			fmt.Printf("Bank population: failure - %v\n", err)
		} else {
			success += 1
		}
	}

	fmt.Printf("Bank population: %d sucess / %d total\n", success, i)
}

func populateATMBanksInternalAlt(session neo4j.SessionWithContext, csvPath string) {

	// 1. Open and read the CSV file
	file, err := os.Open(csvPath + "/atm-bank-internal.csv")
	if err != nil {
		fmt.Printf("could not open CSV file: %v\n", err)
		return
	}
	// closes the file after read from it no matter if there is error or not
	defer file.Close()

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	// Read and discard the header line
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("could not read header from CSV file: %v\n", err)
		return
	}

	query := `
             MATCH (a:ATM {ATM_id: $ATM_id})
             MATCH (b:Bank {code: $code})
             MERGE (a)-[r:BELONGS_TO]->(b);
	`

	i := 0
	success := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		i += 1

		code := row[0]
		ATM_id := row[1]

		params := map[string]interface{}{
			"ATM_id": ATM_id,
			"code":   code,
		}

		err = writeQuery(session, query, params)
		if err != nil {
			fmt.Printf("ATM-Bank-Internal relationships population: failure - %v\n", err)
		} else {
			success += 1
		}
	}

	fmt.Printf("ATM-Bank-Internal relationships population: %d sucess / %d total\n", success, i)
}

func populateATMBanksExternalAlt(session neo4j.SessionWithContext, csvPath string) {

	// 1. Open and read the CSV file
	file, err := os.Open(csvPath + "/atm-bank-external.csv")
	if err != nil {
		fmt.Printf("could not open CSV file: %v\n", err)
		return
	}
	// closes the file after read from it no matter if there is error or not
	defer file.Close()

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	// Read and discard the header line
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("could not read header from CSV file: %v\n", err)
		return
	}

	query := `
             MATCH (a:ATM {ATM_id: $ATM_id})
             MATCH (b:Bank {code: $code})
             MERGE (a)-[r:EXTERNAL]->(b);
	`

	i := 0
	success := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		i += 1

		code := row[0]
		ATM_id := row[1]

		params := map[string]interface{}{
			"ATM_id": ATM_id,
			"code":   code,
		}

		err = writeQuery(session, query, params)
		if err != nil {
			fmt.Printf("ATM-Bank-External relationships population: failure - %v\n", err)
		} else {
			success += 1
		}
	}

	fmt.Printf("ATM-Bank-External relationships population: %d sucess / %d total\n", success, i)
}

func populateCardsAlt(session neo4j.SessionWithContext, csvPath string) {
	// 1. Open and read the CSV file
	file, err := os.Open(csvPath + "/card.csv")
	if err != nil {
		fmt.Printf("could not open CSV file: %v\n", err)
		return
	}
	// closes the file after read from it no matter if there is error or not
	defer file.Close()

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	// Read and discard the header line
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("could not read header from CSV file: %v\n", err)
		return
	}

	query := `
		MERGE (c:Card {
			number_id: $number_id, 
			client_id: $client_id, 
			expiration: $expiration,
			CVC: $CVC,
			extract_limit: $extract_limit, 
			loc_latitude: $loc_latitude, 
			loc_longitude: $loc_longitude
		});
	`
	// To do datetime conversion
	// https://yourbasic.org/golang/format-parse-string-time-date-example/
	const layout = "2006-01-02"

	i := 0
	success := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		i += 1

		number_id := row[0]
		client_id := row[1]

		// date
		expiration, err := time.Parse(layout, row[2])
		if err != nil {
			fmt.Printf("invalid expiration datetime value at row %d: %v\n", i+2, err)
			continue
		}
		// integer
		CVC, err := strconv.Atoi(row[3])
		if err != nil {
			fmt.Printf("invalid CVC value at row %d: %v\n", i+2, err)
			continue
		}
		// float
		extract_limit, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			fmt.Printf("invalid extract_limit value at row %d: %v\n", i+2, err)
			continue
		}
		loc_latitude, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			fmt.Printf("invalid latitude value at row %d: %v\n", i+2, err)
			continue
		}
		loc_longitude, err := strconv.ParseFloat(row[6], 64)
		if err != nil {
			fmt.Printf("invalid longitude value at row %d: %v\n", i+2, err)
			continue
		}

		params := map[string]interface{}{
			"number_id":     number_id,
			"client_id":     client_id,
			"expiration":    expiration,
			"CVC":           CVC,
			"extract_limit": extract_limit,
			"loc_latitude":  loc_latitude,
			"loc_longitude": loc_longitude,
		}

		err = writeQuery(session, query, params)
		if err != nil {
			fmt.Printf("Card population: failure - %v\n", err)
		} else {
			success += 1
		}

	}

	fmt.Printf("Card population: %d sucess / %d total\n", success, i)
}

func populateCardBanksAlt(session neo4j.SessionWithContext, csvPath string) {
	// 1. Open and read the CSV file
	file, err := os.Open(csvPath + "/card-bank.csv")
	if err != nil {
		fmt.Printf("could not open CSV file: %v", err)
		return
	}
	// closes the file after read from it no matter if there is error or not
	defer file.Close()

	// csv reader
	reader := csv.NewReader(bufio.NewReader(file))
	// Read and discard the header line
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("could not read header from CSV file: %v", err)
		return
	}

	query := `
             MATCH (c:Card {number_id: $number_id})
             MATCH (b:Bank {code: $code})
             MERGE (c)-[r:ISSUED_BY]->(b);
	`

	i := 0
	success := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		i += 1

		code := row[0]
		number_id := row[1]

		params := map[string]interface{}{
			"number_id": number_id,
			"code":      code,
		}

		err = writeQuery(session, query, params)
		if err != nil {
			fmt.Printf("Card-Bank relationship population: failure - %v\n", err)
		} else {
			success += 1
		}
	}

	fmt.Printf("Card-Bank relationships population: %d sucess / %d total\n", success, i)

}

func PopulateAlt(csvPath string) {
	fmt.Println("Population of the GDB...")
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	populateATMsAlt(session, csvPath)
	populateBanksAlt(session, csvPath)
	populateATMBanksInternalAlt(session, csvPath)
	populateATMBanksExternalAlt(session, csvPath)
	populateCardsAlt(session, csvPath)
	populateCardBanksAlt(session, csvPath)

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
	err := writeQuery(session, bankUniquenessQuery, nil)
	if err != nil {
		fmt.Printf("Bank uniqueness constraint addition: failure - %v\n", err)
	} else {
		fmt.Println("Bank uniqueness constraint addition: sucess")
	}

	cardUniquenessQuery := `
		CREATE CONSTRAINT number_id IF NOT EXISTS
		FOR (c:Card) REQUIRE c.number_id IS UNIQUE;
	`
	err = writeQuery(session, cardUniquenessQuery, nil)
	if err != nil {
		fmt.Printf("Card uniqueness constraint addition: failure - %v\n", err)
	} else {
		fmt.Println("Card uniqueness constraint addition: sucess")
	}

	ATMUniquenessQuery := `
		CREATE CONSTRAINT ATM_id IF NOT EXISTS
		FOR (a:ATM) REQUIRE a.ATM_id IS UNIQUE;
	`
	err = writeQuery(session, ATMUniquenessQuery, nil)
	if err != nil {
		fmt.Printf("ATM uniqueness constraint addition: failure - %v\n", err)
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
