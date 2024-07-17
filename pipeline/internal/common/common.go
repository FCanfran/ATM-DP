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

// For the volatile subgraph

//const timeTransactionThreshold = 5

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
	g.edges.PushBack(e)
}

// TODO: Idea --> hacer una función que haga el add de una nueva edge
// + el pattern matching check + el update de la lista de tx en base al
// timestamp (eliminando las que estén desfasadas, desde el principio de
// la lista)
// ------------> POR EL MOMENTO, POR SEPARADO.

// Given a certain datetime, it updates the graph, starting from the first
// edge, by eliminating those that are outdated wrt this datetime
// datetime format: DD/MM/YYYY HH:MM:SS
func (g *Graph) Update(timestamp time.Time) {
	// Traverse the list from the beginning and eliminate edges until no
	// outdate is detected
	for eg := g.edges.Front(); eg != nil; eg = eg.Next() {
		eg_val := eg.Value.(Edge) // asserts eg.Value to type Edge
		fmt.Println(eg_val)
		// TODO: HACER ESTO BIEN
		/*
			if eg_val.Tx_end-timestamp >= timeTransactionThreshold {
				g.edges.Remove(eg)
			} else {
				return
			}
		*/
	}
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
