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
const timeTransactionThreshold = 1 * 24 * time.Hour

// TODO: Put this correctly!
const timeFilterThreshold = 10 * 24 * time.Hour

// For the volatile subgraph

// Golang list - doubly linked list
// Graph is a struct that encapsulates a list of edges
type Graph struct {
	edges *list.List
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
		if difference >= timeTransactionThreshold {
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
	eg := g.edges.Back()
	if eg == nil {
		return true
	}
	eg_val := eg.Value.(Edge) // asserts eg.Value to type Edge
	// TODO: tx_start or end?
	difference := timestamp.Sub(eg_val.Tx_start)
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
