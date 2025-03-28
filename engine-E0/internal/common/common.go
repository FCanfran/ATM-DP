package common

import (
	"bufio"
	"container/list"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"pipeline/internal/connection"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/umahmood/haversine"
)

const ChannelSize = 5000

// csv input reading
const ChunkSize = 100

// Parameterized with input execution description file
// ***************************************************************************** //
// max number of cards per filter
var MaxFilterSize int = 4

// diefpy csv result files
var TEST string
var APPROACH string

// stream input file name
var StreamFileName string

// ***************************************************************************** //

// output directory name
var OutDirName string

// scaling factor
var scaleFactor float64 = 1.0

func setOutputDir(name string) {

	// create output dir - if it does not exist
	_, err := os.Stat("./output")
	if os.IsNotExist(err) {
		err = os.Mkdir("./output", 0755)
		CheckError(err)
	}
	CheckError(err)

	// create subdirectory if it does not exist
	OutDirName = "./output/" + name
	_, err = os.Stat(OutDirName)
	if os.IsNotExist(err) {
		err = os.Mkdir(OutDirName, 0755)
		CheckError(err)
	}
	CheckError(err)
}

func ReadExecDescriptionFile(filename string) {
	// csv file content:
	// txFile, test, approach, maxFilterSize, scaleFactor
	file, err := os.Open(filename)
	CheckError(err)
	defer file.Close()
	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read()
	CheckError(err)

	params, err := reader.Read()
	CheckError(err)

	if len(params) != 5 {
		log.Fatalf("Unexpected number of fields in the execution description CSV file\n")
	}

	// txFile
	StreamFileName = params[0]

	// test
	TEST = params[1]
	// approach
	APPROACH = params[2]

	// maxFilterSize
	MaxFilterSize, err = strconv.Atoi(params[3])
	CheckError(err)

	// scaleFactor
	scaleFactorPre, err := strconv.ParseFloat(params[4], 64)
	CheckError(err)
	if scaleFactorPre < 0 || scaleFactorPre > 1 {
		log.Fatalf("Error --- scaleFactor needs to be in [0,1]\n")
	} else {
		scaleFactor = scaleFactorPre
	}

	// create output dir to put the result files
	// obtain the name after input filename
	outdirName := APPROACH
	setOutputDir(outdirName)

	fmt.Println("##############    EXECUTION PARAMETERS    #############")
	fmt.Println("#######################################################")
	fmt.Println(StreamFileName)
	fmt.Println(TEST)
	fmt.Println(APPROACH)
	fmt.Println(MaxFilterSize)
	fmt.Println(scaleFactor)
	fmt.Println("#######################################################")

}

// ***************************************************************************** //

const Time_layout_micro = "2006-01-02 15:04:05.000000"

type TxType uint8

const (
	Withdrawal TxType = 0
	Deposit           = 1
	Inquiry           = 2
	Transfer          = 3
	Other             = 4
)

type EventType uint8

const (
	EdgeStart EventType = 0
	EdgeEnd             = 1
	EOF                 = 2
	LOG                 = 3
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

type Event struct {
	Type      EventType
	E         Edge
	Timestamp time.Duration // internal system timestamp - denotes when the tx arrives to the system
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// CheckFraud() parameters
// Assumption on the maximum speed (km/h) at which the distance between two geographical points
// can be traveled
const maxSpeed = 500 // km/h

// For the volatile subgraph
// Graph is a struct that encapsulates a list of edges: edges
type Graph struct {
	edges *list.List // a list of pointers to edges
}

type Alert struct {
	Label              string        // it can also be set as integer - for each kind of fraud pattern put a int
	Info               string        // optional additional information of the alert to be passed
	Subgraph           Graph         // if desired, if needed later when receiving the alert in the generator
	LastEventTimestamp time.Duration // denotes the internal system timestamp of the last event that composes the alert
}

func (g *Graph) GetEdgeList() *list.List {
	return g.edges
}

// NewGraph creates a new graph
func NewGraph() *Graph {
	g := Graph{edges: list.New()}
	return &g
}

// sets edge e as the new head of the subgraph while erasing the previous
func (g *Graph) NewHead(e Edge) {
	g.edges = list.New()
	g.edges.PushBack(&e)
}

// For adding a general edge or a tx-start edge at the end of the volatile subgraph ds
func (g *Graph) AddEdge(e Edge) {
	g.edges.PushBack(&e) // Adding edge as pointer to the list (list of pointers to edges)
}

// Complete an edge in the subgraph with the tx-end edge
func (g *Graph) CompleteEdge(e Edge) {
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

// ------------------------------------------------------------------------------ //
// obtain Tmin(eg.loc, new_e.loc), returns seconds of time
func obtainTmin(ctx context.Context, session neo4j.SessionWithContext, ATM_id_1 string, ATM_id_2 string) (float64, error) {
	// Connect to the static gdb to obtain the location of the ATMs given the ATM ids
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
	_, distance_km := haversine.Distance(loc1, loc2)

	// t = e / v ---> (km)/(km/h) --> in seconds (*60*60)
	t_min := (distance_km / maxSpeed) * 60 * 60 // in seconds

	return t_min, nil
}

func (g *Graph) CheckFraud(ctx context.Context, session neo4j.SessionWithContext, new_e Edge) (bool, Alert) {

	var fraudAlert Alert // Default 0-value initialization
	fraudIndicator := false

	// 1. Obtain last added edge of the subgraph
	last := g.edges.Back()

	if last != nil {
		last_e := *(last.Value.(*Edge)) // asserts eg.Value to type Edge
		// case new_e.tx_start < last_e.tx_end
		if last_e.Tx_end.IsZero() || new_e.Tx_start.Before(last_e.Tx_end) {
			log.Println("Warning: tx starts before the previous ends! - ", new_e.Number_id)
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			fmt.Println("Warning: tx starts before the previous ends! - ", new_e.Number_id)
			PrintEdge("Fraud pattern with: ", last_e)
		} else {
			if last_e.ATM_id != new_e.ATM_id {
				// time feasibility check: (new_e.tx_start - last_e.tx_end) < Tmin(last_e.loc, new_e.loc)
				// obtain Tmin(last_e.loc, new_e.loc)
				t_min, err := obtainTmin(ctx, session, last_e.ATM_id, new_e.ATM_id)
				CheckError(err)
				// scale factor
				t_min = t_min * scaleFactor
				t_diff := (new_e.Tx_start.Sub(last_e.Tx_end)).Seconds()
				if t_diff < t_min {
					// create alert
					subgraph := NewGraph()
					subgraph.AddEdge(last_e)
					subgraph.AddEdge(new_e)
					fraudAlert.Label = "1"
					fraudAlert.Info = "fraud pattern"
					fraudAlert.Subgraph = *subgraph
					fraudIndicator = true
				}
			}
		}
	}

	return fraudIndicator, fraudAlert

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

// Constructs the Edge Event from a transaction csv row
func ReadEdge(tx []string) Event {

	var r Event
	// id
	tx_id_64, err := strconv.ParseInt(tx[0], 10, 32) // 10: base (decimal) & 32: bit-size (int32)
	CheckError(err)
	tx_id := int32(tx_id_64) // still the type returned is int64 -> convert to int32

	// type
	var tx_type TxType
	tx_type_64, err := strconv.ParseInt(tx[3], 10, 8) // int8
	CheckError(err)
	if tx_type_64 < 0 || tx_type_64 > 3 {
		tx_type = Other
	} else {
		tx_type = TxType(tx_type_64)
	}

	// start
	tx_start, err := time.Parse(Time_layout_micro, tx[4])
	CheckError(err)

	// end
	// Check if tx_end field is empty
	var tx_end time.Time
	if tx[5] != "" {
		tx_end, err = time.Parse(Time_layout_micro, tx[5])
		CheckError(err)
		r.Type = EdgeEnd
	} else {
		r.Type = EdgeStart // tx_end field is empty
	}

	var tx_amount_32 float32
	if tx[6] != "" {
		tx_amount, err := strconv.ParseFloat(tx[6], 32)
		CheckError(err)
		tx_amount_32 = float32(tx_amount)
	}

	edge := Edge{
		Number_id: tx[1],
		ATM_id:    tx[2],
		Tx_id:     tx_id,
		Tx_type:   tx_type,
		Tx_start:  tx_start,
		Tx_end:    tx_end,
		Tx_amount: tx_amount_32,
	}
	r.E = edge
	return r
}

// Print a subgraph
func (g *Graph) Print() {
	if g.edges.Front() != nil {
		card_id := g.edges.Front().Value.(*Edge).Number_id
		fmt.Println("subgraph: ", card_id)
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
		card_id := g.edges.Front().Value.(*Edge).Number_id
		fmt.Fprintln(file, "subgraph: ", card_id)
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
		card_id := g.edges.Front().Value.(*Edge).Number_id
		fmt.Println("subgraph: ", card_id)
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
		fmt.Printf("%d,%s,%s,%d,%s,%s,%.2f\n",
			e.Tx_id,
			e.Number_id,
			e.ATM_id,
			e.Tx_type,
			e.Tx_start.Format(Time_layout_micro),
			e.Tx_end.Format(Time_layout_micro),
			e.Tx_amount)
	} else {
		fmt.Printf("%s  %d,%s,%s,%d,%s,%s,%.2f\n",
			msg,
			e.Tx_id,
			e.Number_id,
			e.ATM_id,
			e.Tx_type,
			e.Tx_start.Format(Time_layout_micro),
			e.Tx_end.Format(Time_layout_micro),
			e.Tx_amount)
	}
}

func PrintEdgeCompleteToFile(msg string, e Edge, file *os.File) {
	if msg == "" {
		fmt.Fprintf(file, "%d,%s,%s,%d,%s,%s,%.2f\n",
			e.Tx_id,
			e.Number_id,
			e.ATM_id,
			e.Tx_type,
			e.Tx_start.Format(Time_layout_micro),
			e.Tx_end.Format(Time_layout_micro),
			e.Tx_amount)
	} else {
		fmt.Fprintf(file, "%s:   %d,%s,%s,%d,%s,%s,%.2f\n",
			msg,
			e.Tx_id,
			e.Number_id,
			e.ATM_id,
			e.Tx_type,
			e.Tx_start.Format(Time_layout_micro),
			e.Tx_end.Format(Time_layout_micro),
			e.Tx_amount)
	}
}

func PrintAlertVerbose(alert Alert, timestamp time.Duration, alertCount int) {

	fmt.Printf("Alert - label: %s, info: %s, timestamp: %v, numAlert: %d\n", alert.Label, alert.Info, timestamp, alertCount)
	switch alert.Label {
	case "0", "1":
		alert.Subgraph.Print()
	}
	fmt.Println("______________________________________________________________________________")
}

func PrintAlertOnFileVerbose(alert Alert, responseTime time.Duration, alertCount int, file *os.File) {
	fmt.Fprintf(file, "Alert - label: %s, info: %s, responseTime: %v, numAlert: %d\n", alert.Label, alert.Info, responseTime, alertCount)
	switch alert.Label {
	case "0", "1":
		alert.Subgraph.PrintToFile(file)
	}
}

func PrintAlertOnFile(alert Alert, file *os.File) {
	fmt.Fprintf(file, "Alert - label: %s\n", alert.Label)
	switch alert.Label {
	case "0", "1":
		alert.Subgraph.PrintToFile(file)
	}
}

func PrintEventOnFile(e Event, file *os.File) {

	switch e.Type {
	case EOF:
		fmt.Fprintf(file, "Event - type: EOF\n")
	case LOG:
		fmt.Fprintf(file, "Event - type: LOG\n")
		fmt.Fprintf(file, "Event - edge: ")
		edge := e.E
		// transaction_id,number_id,ATM_id,transaction_start,transaction_end,transaction_amount
		out_string := fmt.Sprintf("%d,%s,%s,%d,%s,%s,%.2f\n",
			edge.Tx_id,
			edge.Number_id,
			edge.ATM_id,
			edge.Tx_type,
			edge.Tx_start.Format(Time_layout_micro),
			edge.Tx_end.Format(Time_layout_micro),
			edge.Tx_amount)
		file.WriteString(out_string)
	default:
		fmt.Fprintf(file, "Event - type: OTHER\n")
		fmt.Fprintf(file, "Event - edge: ")
		edge := e.E
		// transaction_id,number_id,ATM_id,transaction_start,transaction_end,transaction_amount
		out_string := fmt.Sprintf("%d,%s,%s,%d,%s,%s,%.2f\n",
			edge.Tx_id,
			edge.Number_id,
			edge.ATM_id,
			edge.Tx_type,
			edge.Tx_start.Format(Time_layout_micro),
			edge.Tx_end.Format(Time_layout_micro),
			edge.Tx_amount)
		file.WriteString(out_string)
	}

}

func PrintAlertOnResultsTrace(timestamp time.Duration, responseTime time.Duration, alertCount int, csv_writer *csv.Writer) {
	dataRow := []string{
		TEST,                     // test (stream kind)
		APPROACH,                 // approach (num cores & num filters)
		strconv.Itoa(alertCount), // answer
		strconv.FormatFloat(timestamp.Seconds(), 'f', 2, 64),                 // time (in seconds)
		strconv.FormatFloat(float64(responseTime.Nanoseconds()), 'f', 2, 64), // in nanoseconds
	}

	err := csv_writer.Write(dataRow)
	CheckError(err)

	csv_writer.Flush() // Ensure data is written to file
}

func PrintMetricsResults(timeFirst time.Duration, timeLast time.Duration, alertCount int, csv_writer *csv.Writer) {
	dataRow := []string{
		TEST,     // test
		APPROACH, // approach
		strconv.FormatFloat(timeFirst.Seconds(), 'f', 2, 64),  // tfft time (in seconds)
		strconv.FormatFloat(timeLast.Seconds(), 'f', 2, 64),   // totaltime time (in seconds)
		strconv.FormatInt(time.Duration(0).Nanoseconds(), 10), // response time (in nanoseconds) - draft filling value - calculate afterwards!
		strconv.Itoa(alertCount),                              // comp
	}

	err := csv_writer.Write(dataRow)
	CheckError(err)

	csv_writer.Flush() // Ensure data is written to file
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
