package common

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"os"
	"pipeline/internal/connection"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/umahmood/haversine"
)

// https://yourbasic.org/golang/format-parse-string-time-date-example/
const Time_layout = "2006-01-02 15:04:05"

// An Edge = Transaction: Card ---> ATM
// It is an edge of the volatile subgraph
type Edge struct {
	Number_id string    // Card id
	ATM_id    string    // ATM id
	Tx_id     int32     // transaction id
	Tx_start  time.Time // transaction start date time (DD/MM/YYYY HH:MM:SS)
	Tx_end    time.Time // transaction end date time (DD/MM/YYYY HH:MM:SS)
	Tx_amount float32   // transaction amount
	// Tx_type: type of tx (NOTE: for the moment only withdrawals are considered)
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// FUTURE: For the multiple window support - for the moment: single window support
// ------------------------------------------------------------------------------ //
// TODO: Put this correctly!, for the moment the diff is 24h
// In Duration format
const timeTxThreshold = 1 * 24 * time.Hour

// TODO: Put this correctly!
const timeFilterThreshold = 2 * 24 * time.Hour

// ------------------------------------------------------------------------------ //

// CheckFraud() parameters
// Assumption on the maximum speed (km/h) at which the distance between two geographical points
// can be traveled
const maxSpeed = 500 // km/h

// ------------------------------------------------------------------ //

// For the volatile subgraph

// Golang list - doubly linked list
// Graph is a struct that encapsulates a list of edges: edges
// last_timestamp: to save the last timestamp of the filter / subgraph
type Graph struct {
	last_timestamp time.Time // Tx_end of the last edge to have been added to the subgraph
	edges          *list.List
}

type Alert struct {
	Label         string // it can also be set as integer - for each kind of fraud pattern put a int
	Info          string // optional additional information of the alert to be passed
	Subgraph      Graph  // if desired, if needed later when receiving the alert in the generator
	AnomalousEdge Edge   // the anomalous tx itself
}

// NewGraph creates a new graph
func NewGraph() *Graph {
	g := Graph{edges: list.New()}
	return &g
}

// Appends a new edge at the end of the list
func (g *Graph) AddAtEnd(e Edge) {
	//fmt.Println(":::: addition ::::")
	g.edges.PushBack(e)
	g.last_timestamp = e.Tx_end
}

// FUTURE: For the multiple window support - for the moment: single window support
// ------------------------------------------------------------------------------ //
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

// ------------------------------------------------------------------------------ //

// obtain Tmin(eg.loc, new_e.loc), returns seconds of time
func obtainTmin(ctx context.Context, session neo4j.SessionWithContext, ATM_id_1 string, ATM_id_2 string) (int, error) {
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
	_, distance_km := haversine.Distance(loc1, loc2)
	fmt.Println("Kilometers:", distance_km)

	// t = e / v ---> (km)/(km/h) --> in seconds (*60*60)
	t_min := (distance_km / maxSpeed) * 60 * 60 // in seconds

	return int(t_min), nil
}

func (g *Graph) CheckFraud(new_e Edge) (bool, *Graph, Edge) {

	fmt.Println("-------------- CHECKFRAUD()--------------")
	var isFraud bool = false
	var subgraph *Graph = nil
	var anomalousEdge Edge
	// New root context for the connections to the gdb that are going to be done here
	context := context.Background()
	// 0. Open a session to connect to the gdb
	session := connection.CreateSession(context)
	defer connection.CloseSession(context, session)

	// 1. Obtain last added edge of the subgraph
	for prev := g.edges.Back(); prev != nil; prev = prev.Next() {

		prev_e := prev.Value.(Edge) // asserts eg.Value to type Edge

		// Case new_e.tx_start < prev_e.tx_end -> it can't happen that a transaction starts before the previous is finished
		if new_e.Tx_start.Before(prev_e.Tx_end) {
			fmt.Println("tx starts before the previous ends!")
			// TODO: It is a TRUE fraud, but not of this kind! - other kind
			// do not consider it here so far
			isFraud = false
			// print fraud pattern with this edge
			PrintEdge("Fraud pattern with: ", prev_e)
			continue
		}

		if prev_e.ATM_id == new_e.ATM_id {
			continue // no fraud with current edge, go to check next edge
		}

		// != ATM_id

		// time feasibility check: (new_e.tx_start - prev_e.tx_end) < Tmin(prev_e.loc, new_e.loc)
		// obtain Tmin(eg.loc, new_e.loc)
		t_min, err := obtainTmin(context, session, prev_e.ATM_id, new_e.ATM_id)
		CheckError(err)

		t_diff := int((new_e.Tx_start.Sub(prev_e.Tx_end)).Seconds())

		fmt.Println("t_min", t_min)
		fmt.Println("t_diff", t_diff)

		if t_diff < t_min {
			isFraud = true
			// print fraud pattern with this edge
			PrintEdge("Fraud pattern with: ", prev_e)
			// subgraph
			subgraph = NewGraph()
			subgraph.AddAtEnd(prev_e)
			subgraph.AddAtEnd(new_e)
			// anomalous edge
			anomalousEdge = new_e
		}
	}

	// if the subgraph was empty, then isFraud is false...
	return isFraud, subgraph, anomalousEdge

}

// Print a subgraph
func (g *Graph) Print() {
	if g.edges.Front() != nil {
		filter_id := g.edges.Front().Value.(Edge).Number_id
		fmt.Println("subgraph: ", filter_id)
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
		for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
			eg_val := eg.Value.(Edge)
			fmt.Println(eg_val)
		}
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
	}
}

// Print a subgraph - only the tx ids
func (g *Graph) PrintIds() {
	if g.edges.Front() != nil {
		filter_id := g.edges.Front().Value.(Edge).Number_id
		fmt.Println("subgraph: ", filter_id)
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
		for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
			eg_val := eg.Value.(Edge)
			fmt.Println(eg_val.Tx_id)
		}
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
	}
}

// Generic Functions -----------------------------

func PrintEdge(msg string, e Edge) {
	if msg == "" {
		fmt.Printf("%d, %s -> %s\n", e.Tx_id, e.Number_id, e.ATM_id)
	} else {
		fmt.Printf("%s  %d, %s -> %s\n", msg, e.Tx_id, e.Number_id, e.ATM_id)
	}
}

func PrintEdgeComplete(msg string, e Edge) {
	if msg == "" {
		fmt.Printf("%d,%s,%s,%s,%s,%.2f\n",
			e.Tx_id,
			e.Number_id,
			e.ATM_id,
			e.Tx_start.Format(Time_layout),
			e.Tx_end.Format(Time_layout),
			e.Tx_amount)
	} else {
		fmt.Printf("%s:   %d,%s,%s,%s,%s,%.2f\n",
			msg,
			e.Tx_id,
			e.Number_id,
			e.ATM_id,
			e.Tx_start.Format(Time_layout),
			e.Tx_end.Format(Time_layout),
			e.Tx_amount)
	}
}

func PrintAlertVerbose(alert Alert) {
	fmt.Printf("Alert!: %s, %s\n", alert.Label, alert.Info)

	switch alert.Label {
	case "1":
		fmt.Print("Anomalous tx: ")
		PrintEdge("", alert.Subgraph.edges.Back().Value.(Edge))
		fmt.Println("....................")
		alert.Subgraph.Print()

	}
	fmt.Println("______________________________________________________________________________")
}

// So far, to print the id of the anomalous tx on a dedicated log file for each kind of fraud
func PrintAlertOnFile(alert Alert, file *os.File) {
	switch alert.Label {
	case "1":
		// get the id of the anomalous tx
		file.WriteString(strconv.Itoa(int(alert.AnomalousEdge.Tx_id)) + "\n")
	}
}

// TODO: For the moment the events are ONLY edges
func PrinteEventOnFile(e Edge, file *os.File) {
	// transaction_id,number_id,ATM_id,transaction_start,transaction_end,transaction_amount
	out_string := fmt.Sprintf("%d,%s,%s,%s,%s,%.2f\n",
		e.Tx_id,
		e.Number_id,
		e.ATM_id,
		e.Tx_start.Format(Time_layout),
		e.Tx_end.Format(Time_layout),
		e.Tx_amount)
	file.WriteString(out_string)

}

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
