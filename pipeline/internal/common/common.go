package common

import (
	"container/list"
	"fmt"
	"log"
	"time"
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
func obtainTmin() float32 {
	// Connect to the static gdb to obtain the location of the ATMs given the ATM ids
}

func (g *Graph) CheckFraud(new_e Edge) bool {
	// NOTE: Initial version - pattern 1 - easy approach (only check with the last added edge of the subgraph)

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
	obtainTmin()
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
