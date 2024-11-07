package common

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"os"
	"pipeline/internal/connection"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/umahmood/haversine"
)

const ChannelSize = 5000

// https://yourbasic.org/golang/format-parse-string-time-date-example/
const Time_layout = "2006-01-02 15:04:05"

type TxType uint8

const (
	Withdrawal TxType = 0
	Deposit           = 1
	Inquiry           = 2
	Transfer          = 3
	Other             = 4
)

type Operation uint8

const (
	TX  Operation = 1
	EOF           = 2
)

// An Edge = Transaction: Card ---> ATM
// It is an edge of the volatile subgraph
type Edge struct {
	Number_id string    // Card id
	ATM_id    string    // ATM id
	Tx_id     int32     // transaction id
	Tx_type   TxType    // transaction type (withdrawal/deposit/inquiry/transfer)
	Tx_start  time.Time // transaction start date time (DD/MM/YYYY HH:MM:SS)
	Tx_end    time.Time // transaction end date time (DD/MM/YYYY HH:MM:SS)
	Tx_amount float32   // transaction amount
}

type Request struct {
	Op Operation
	E  Edge
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
	last_timestamp time.Time
	edges          *list.List // a list of pointers to edges
}

type Alert struct {
	Label    string // it can also be set as integer - for each kind of fraud pattern put a int
	Info     string // optional additional information of the alert to be passed
	Subgraph Graph  // if desired, if needed later when receiving the alert in the generator
}

// NewGraph creates a new graph
func NewGraph() *Graph {
	g := Graph{edges: list.New()}
	return &g
}

// For adding a general edge or a tx-start edge at the end of the volatile subgraph ds
func (g *Graph) AddEdge(e Edge) {
	fmt.Println(":::: AddEdge ::::")
	g.edges.PushBack(&e) // Adding edge as pointer to the list (list of pointers to edges)
	// TODO: Needs to be updated?
	// g.last_timestamp = e.Tx_start
}

// Complete an edge in the subgraph with the tx-end edge
func (g *Graph) CompleteEdge(e Edge) {
	fmt.Println(":::: CompleteEdge ::::")
	// Get the last edge of the list and complete it
	// we are getting a reference of the object, so any change directly modifies it
	prev := g.edges.Back()
	// Security check
	if prev != nil {
		edge := prev.Value.(*Edge)
		if edge.Tx_id == e.Tx_id {
			// Complete the edge with the tx-end information
			edge.Tx_end = e.Tx_end
			edge.Tx_amount = e.Tx_amount
		} else {
			log.Println("Warning: AddEdge ->  possible overlapping: a tx-end of a different tx-id was received before the previous tx was closed")
		}
	} else {
		log.Println("Warning: AddEdge -> a tx-end was tryied to be added on a empty subgraph")
	}
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

// TODO: CHECK -> When do we stop checking, when we detect the first anomalous scenario,
// or when we have traversed the full list of edges - all possible scenarios...
// -------> for the moment, stop when we discover the first. Only 1 at most is reported.º
// Parameters:
// - new_e: the new edge that we check the FP against
// Returns:
// - bool: indicating the presence of a positive Alert (true) or not (false)
// - Alert: the Alert itself, only in the case it is true. Empty if false.
func (g *Graph) CheckFraud(new_e Edge) (bool, Alert) {

	fmt.Println("-------------- CHECKFRAUD()--------------")
	// New root context for the connections to the gdb that are going to be done here
	context := context.Background()
	// 0. Open a session to connect to the gdb
	session := connection.CreateSession(context)
	defer connection.CloseSession(context, session)

	fraudIndicator := false
	var fraud1Alert Alert // Default 0-value initialization

	// 1. Obtain last added edge of the subgraph
	// TODO: FIX ---> I think the backward traversal is incorrect!
	for prev := g.edges.Back(); prev != nil && !fraudIndicator; prev = prev.Prev() {

		prev_e := *(prev.Value.(*Edge)) // asserts eg.Value to type Edge

		// Case new_e.tx_start < prev_e.tx_end -> it can't happen that a transaction starts before the previous is finished
		if new_e.Tx_start.Before(prev_e.Tx_end) {
			fmt.Println("tx starts before the previous ends!")
			// TODO: It is a TRUE fraud, but not of this kind! - other kind
			// do not consider it here so far
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
			// print fraud pattern with this edge
			PrintEdge("Fraud pattern with: ", prev_e)
			// subgraph
			subgraph := NewGraph()
			subgraph.AddEdge(prev_e)
			subgraph.AddEdge(new_e)
			fmt.Println("TRUE FP1: ")
			subgraph.Print()
			fmt.Println("...........................................................")
			// Construct the corresponding Alert properly
			fraud1Alert.Label = "1"
			fraud1Alert.Info = "fraud pattern"
			fraud1Alert.Subgraph = *subgraph
			fraudIndicator = true
		} else {
			fmt.Println("FALSE FP1")
		}
	}

	return fraudIndicator, fraud1Alert
}

// Returns true if the tx edge is start.
// Based on the tx_end property -> isZero in that case.
func (e Edge) IsStart() bool {
	if e.Tx_end.IsZero() {
		return true
	} else {
		return false
	}
}

// Print a subgraph
func (g *Graph) Print() {
	if g.edges.Front() != nil {
		filter_id := g.edges.Front().Value.(*Edge).Number_id
		fmt.Println("subgraph: ", filter_id)
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
		for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
			eg_val := eg.Value.(*Edge)
			PrintEdgeComplete("", *eg_val)
		}
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
	}
}

func (g *Graph) PrintToFile(file *os.File) {
	if g.edges.Front() != nil {
		filter_id := g.edges.Front().Value.(*Edge).Number_id
		fmt.Fprintln(file, "subgraph: ", filter_id)
		fmt.Fprintln(file, "- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
		for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
			eg_val := eg.Value.(*Edge)
			PrintEdgeCompleteToFile("", *eg_val, file)
		}
		fmt.Fprintln(file, "- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
	}
}

// Print a subgraph - only the tx ids
func (g *Graph) PrintIds() {
	if g.edges.Front() != nil {
		filter_id := g.edges.Front().Value.(*Edge).Number_id
		fmt.Println("subgraph: ", filter_id)
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
		for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
			eg_val := eg.Value.(*Edge)
			fmt.Println(eg_val.Tx_id)
		}
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
	}
}

// ------------------------------ Generic Functions ------------------------------

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

func PrintEdgeCompleteToFile(msg string, e Edge, file *os.File) {
	if msg == "" {
		fmt.Fprintf(file, "%d,%s,%s,%s,%s,%.2f\n",
			e.Tx_id,
			e.Number_id,
			e.ATM_id,
			e.Tx_start.Format(Time_layout),
			e.Tx_end.Format(Time_layout),
			e.Tx_amount)
	} else {
		fmt.Fprintf(file, "%s:   %d,%s,%s,%s,%s,%.2f\n",
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
		alert.Subgraph.Print()
	}
	fmt.Println("______________________________________________________________________________")
}

// So far, to print the anomalous tx subgraph on a dedicated log file for each kind of fraud
func PrintAlertOnFile(alert Alert, file *os.File) {
	switch alert.Label {
	case "1":
		alert.Subgraph.PrintToFile(file)
	}
}

// TODO: For the moment the events are ONLY edges
func PrintEventOnFile(e Edge, file *os.File) {
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

// --------------------------------------------------------------------------------------

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
