package common

import (
	"log"
	"time"
)

// An Edge = Transaction: Card ---> ATM 
// It is an edge of the volatile subgraph
type Edge struct {
	Number_id string 	// Card id
	ATM_id string		// ATM id
	Tx_id int64			// transaction id
	Tx_start time.Time	// transaction start date time (DD/MM/YYYY HH:MM:SS)
	Tx_end time.Time	// transaction end date time (DD/MM/YYYY HH:MM:SS)
	Tx_amount float32	// transaction amount 
}

// ------------------------------------------------------------------------------
// TODO: For the volatile subgraph?

type Graph []Edge

/*
func (g *Graph) Delete(e Edge) {
	// For eficiency porpuse, instead of deleting
	// the edge, we substitute it with the last
	// and shrink it. If the order matters, another
	// data structure will be more suitable
	for i, edge := range *g {
		if edge.Card == e.Card && edge.ATM == e.ATM {
			(*g)[i] = (*g)[len(*g)-1]
			(*g) = (*g)[:len(*g)-1]
			return
		}
	}
}

// If the edge already exists, then it updates its Time
// Otherwise it is added to the graph
func (g *Graph) InsertUpdate(e Edge) {
	for _, edge := range *g {
		if edge.Card == e.Card && edge.ATM == e.ATM {
			edge.Time = e.Time
			return
		}
	}
	(*g) = append(*g, e)
}
*/
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
