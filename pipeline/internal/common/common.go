package common

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"pipeline/internal/connection"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/umahmood/haversine"
)

// An Edge = Transaction: Card ---> ATM
// It is an edge of the volatile subgraph
type Edge struct {
	Number_id string    // Card id
	ATM_id    string    // ATM id
	Tx_id     int64     // transaction id
	Tx_start  time.Time // transaction start date time (DD/MM/YYYY HH:MM:SS)
	Tx_end    time.Time // transaction end date time (DD/MM/YYYY HH:MM:SS)
	Tx_amount float32   // transaction amount
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// TODO: Put this correctly!, for the moment the diff is 24h
// In Duration format
const timeTxThreshold = 1 * 24 * time.Hour

// TODO: Put this correctly!
const timeFilterThreshold = 2 * 24 * time.Hour

// For the volatile subgraph

// Golang list - doubly linked list
// Graph is a struct that encapsulates a list of edges: edges
// last_timestamp: to save the last timestamp of the filter / subgraph
type Graph struct {
	last_timestamp time.Time // Tx_end of the last edge to have been added to the subgraph
	edges          *list.List
}

// NewGraph creates a new graph
func NewGraph() *Graph {
	g := Graph{edges: list.New()}
	return &g
}

// Appends a new edge at the end of the list
func (g *Graph) AddAtEnd(e Edge) {
	fmt.Println(":::: addition ::::")
	g.edges.PushBack(e)
	g.last_timestamp = e.Tx_end
}

// TODO: Idea --> hacer una función que haga el add de una nueva edge
// + el pattern matching check + el update de la lista de tx en base al
// timestamp (eliminando las que estén desfasadas, desde el principio de
// la lista)
// ------------> POR EL MOMENTO, POR SEPARADO.

// Given a certain datetime, it updates the graph, starting from the first
// edge, by eliminating those that are outdated wrt this datetime
// - datetime format: DD/MM/YYYY HH:MM:SS
func (g *Graph) Update(timestamp time.Time) {
	fmt.Println(":::: update ::::")
	// Traverse the list from the beginning and eliminate edges until no
	// outdate is detected
	eg := g.edges.Front()
	for eg != nil {
		eg_val := eg.Value.(Edge) // asserts eg.Value to type Edge
		difference := timestamp.Sub(eg_val.Tx_end)
		if difference >= timeTxThreshold {
			// Keep the next before deleting the current, so that we can have
			// the next of the current after the removal
			eg_next := eg.Next()
			g.edges.Remove(eg)
			eg = eg_next
		} else {
			// at the time that we find the first edge which is not
			// outdated, we stop, since for sure the next ones are
			// also not outdated (we are assuming that the tx are
			// received ordered in time...)
			return
		}
	}
}

// Filter timeout check: test if the filter has to die (with the last edge
// of the volatile subgraph and a timestamp), if the time difference is
// greater than timeFilterThreshold returns true, otherwise false
// TODO: Again this is assuming that the tx are ordered in time!!!
// otherwise we will have to find the most recent tx in time
func (g *Graph) CheckFilterTimeout(timestamp time.Time) bool {
	fmt.Println(":::: checkFilterTimeout ::::")
	difference := timestamp.Sub(g.last_timestamp)
	return (difference >= timeFilterThreshold)
}

// Delete a specific edge of the graph
// - Locate by tx id
func (g *Graph) Delete(e Edge) {

	// Locate the element, and then remove it
	for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
		eg_val := eg.Value.(Edge) // asserts eg.Value to type Edge
		if eg_val.Tx_id == e.Tx_id {
			g.edges.Remove(eg)
			return
		}
	}
}

// obtain Tmin(eg.loc, new_e.loc)
func obtainTmin(ctx context.Context, session neo4j.SessionWithContext, ATM_id_1 string, ATM_id_2 string) (float32, error) {
	// Connect to the static gdb to obtain the location of the ATMs given the ATM ids
	// TODO: Use Indexes for Performance
	// Ensure that the ATM_id field is indexed if you are performing many lookups based on this property.
	// While this is not a different query form, indexing helps improve the performance of queries that filter on this property.
	getATMLocationQuery := `MATCH (a:ATM) WHERE a.ATM_id = $ATM_id RETURN a.loc_latitude AS loc_latitude, a.loc_longitude AS loc_longitude`

	processCoordinates := func(result neo4j.ResultWithContext) (any, error) {

		var loc Coordinates

		for result.Next(ctx) {
			record := result.Record()

			loc_latitude, found_lat := record.Get("loc_latitude")
			loc_longitude, found_long := record.Get("loc_longitude")
			if found_lat && found_long {
				if lat, ok := loc_latitude.(float64); ok {
					loc.Latitude = lat
				} else {
					return loc, fmt.Errorf("expected loc_latitude to be float64, but got %T", loc_latitude)
				}
				if long, ok := loc_longitude.(float64); ok {
					loc.Longitude = long
				} else {
					return loc, fmt.Errorf("expected loc_longitude to be float64, but got %T", loc_longitude)
				}
			} else {
				return loc, fmt.Errorf("latitude or longitude not found in record")
			}
		}

		// Check for errors after processing the results
		if err := result.Err(); err != nil {
			return loc, err
		}
		return loc, nil
	}

	params := map[string]any{
		"ATM_id": ATM_id_1,
	}
	result1, err := connection.ReadQuery(ctx, session, getATMLocationQuery, params, processCoordinates)
	CheckError(err)
	var location1 Coordinates
	// Assert to type Coordinates
	if location, ok := result1.(Coordinates); ok {
		location1 = location
	}

	params["ATM_id"] = ATM_id_2
	result2, err := connection.ReadQuery(ctx, session, getATMLocationQuery, params, processCoordinates)
	CheckError(err)
	var location2 Coordinates
	// Assert to type Coordinates
	if location, ok := result2.(Coordinates); ok {
		location2 = location
	}

	// Calculate the distance between the two locations
	loc1 := haversine.Coord{Lat: location1.Latitude, Lon: location1.Longitude}
	loc2 := haversine.Coord{Lat: location2.Latitude, Lon: location2.Longitude}
	fmt.Println(loc1)
	fmt.Println(loc2)
	_, km := haversine.Distance(loc1, loc2)
	fmt.Println("Kilometers:", km)

	return 0, nil
}

func (g *Graph) CheckFraud(new_e Edge) bool {
	// NOTE: Initial version - pattern 1 - easy approach (only check with the last added edge of the subgraph)
	fmt.Println("-------------- CHECKFRAUD()--------------")
	// New root context for the connections to the gdb that are going to be done here
	context := context.Background()
	// 0. Open a session to connect to the gdb
	session := connection.CreateSession(context)
	defer connection.CloseSession(context, session)

	// 1. Obtain last added edge of the subgraph
	eg := g.edges.Back()
	// 2. Check if the subgraph is empty - no fraud
	if eg == nil {
		return false
	}
	eg_val := eg.Value.(Edge) // asserts eg.Value to type Edge

	if eg_val.ATM_id == new_e.ATM_id {
		return false
	}

	// != ATM_id

	// time feasibility check: (new_e.tx_start - eg.tx_end) < Tmin(eg.loc, new_e.loc)
	// obtain Tmin(eg.loc, new_e.loc)
	obtainTmin(context, session, eg_val.ATM_id, new_e.ATM_id)
	// TODO: Do this properly! REMOVE
	return false
}

// Print a subgraph
func (g *Graph) Print() {
	if g.edges.Front() != nil {
		filter_id := g.edges.Front().Value.(Edge).Number_id
		fmt.Println("subgraph: ", filter_id)
		fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
		for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
			eg_val := eg.Value.(Edge)
			fmt.Println(eg_val)
		}
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	}
}

// Print a subgraph - only the tx ids
func (g *Graph) PrintId() {
	if g.edges.Front() != nil {
		filter_id := g.edges.Front().Value.(Edge).Number_id
		fmt.Println("subgraph: ", filter_id)
		fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
		for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
			eg_val := eg.Value.(Edge)
			fmt.Println(eg_val.Tx_id)
		}
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	}
}

// Generic Functions -----------------------------

func CheckError(e error) {
	if e != nil {
		ChangeLogPrefiX()
		log.Fatalf("Fatal error --- %s\n", e.Error())
	}
}

func ChangeLogPrefiX() {
	// Set microseconds and full PATH of source code in logs
	log.SetFlags(log.Lmicroseconds | log.Llongfile)
}
