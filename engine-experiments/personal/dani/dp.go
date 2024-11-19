package dp

import (
	"bufio"
	cmn "dp_mst/internal/common"
	"dp_mst/internal/mst"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const channelSize = 5000

type Msg struct {
	Graph   cmn.Graph
	Updated bool
}

type in_comm struct {
	Req   <-chan cmn.Request
	Graph <-chan Msg
}

type out_comm struct {
	Req   chan<- cmn.Request
	Graph chan<- Msg
}

func input(istream string, out out_comm, Fsize int, inputSync <-chan struct{}) {
	file, err := os.Open(istream + ".requests")
	cmn.CheckError(err)
	defer file.Close()
	cmn.CheckError(err)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	finish := false

	var r cmn.Request
	for !finish {
		select {
		case x, _ := <-c:
			out.Req <- cmn.Request{Op: cmn.EOF, E: cmn.Edge{X: -1, Y: -1, W: 0}}
			empty := make(cmn.Graph, 0)
			out.Graph <- Msg{empty, false}
			fmt.Println("Signal ", x, " received. Finishing work!!")
			finish = true
			continue
		default:
			// No signal, so continue
		}

		scanner.Scan()
		op := scanner.Text()
		err = scanner.Err()
		cmn.CheckError(err)
		if err == io.EOF || op == "" {
			out.Req <- cmn.Request{Op: cmn.EOF, E: cmn.Edge{X: -1, Y: -1, W: 0}}
			empty := make(cmn.Graph, 0)
			out.Graph <- Msg{empty, false}
			fmt.Println("Err := ", err, " Op :=", op)
			break
		}

		op_int, err := strconv.ParseInt(op, 10, 64)
		if err != nil {
			out.Req <- cmn.Request{Op: cmn.EOF, E: cmn.Edge{X: -1, Y: -1, W: 0}}
			empty := make(cmn.Graph, 0)
			out.Graph <- Msg{empty, false}
			break
		}
		cmn.CheckError(err)

		r.Op = cmn.Operation(op_int)
		//fmt.Println("Readed operation: ", r.Op)

		if r.Op == cmn.KMST || r.Op == cmn.GraphOp || r.Op == cmn.EOF || r.Op == cmn.LoadState || r.Op == cmn.SaveState || r.Op == cmn.CurrTime {
			r.E = cmn.Edge{X: -1, Y: -1, W: 0}
		} else {
			scanner.Scan()
			node1 := scanner.Text()
			scanner.Scan()
			node2 := scanner.Text()
			scanner.Scan()
			weight := scanner.Text()
			err = scanner.Err()
			cmn.CheckError(err)

			r.E.X, err = strconv.ParseInt(node1, 10, 32)
			r.E.Y, err = strconv.ParseInt(node2, 10, 32)
			r.E.W, err = strconv.ParseFloat(weight, 64)
		}

		r.Normalize()

		out.Req <- r
		if r.Op == cmn.KMST || r.Op == cmn.GraphOp || r.Op == cmn.EOF || r.Op == cmn.LoadState || r.Op == cmn.SaveState {
			empty := make(cmn.Graph, 0)
			out.Graph <- Msg{empty, false}
		}

		if r.Op == cmn.LoadState {
			<-inputSync
		}

		if r.Op == cmn.Insert || r.Op == cmn.Delete {
			out.Req <- cmn.Request{Op: cmn.KMST, E: cmn.Edge{X: -1, Y: -1, W: 0}}
			empty := make(cmn.Graph, 0)
			out.Graph <- Msg{empty, false}
		}

		if r.Op == cmn.EOF {
			break
		}
	}
	fmt.Println("Input finished")
	close(out.Req)
	close(out.Graph)
}

func output(istream string, in in_comm, end chan<- struct{}) {
	var timestamps []time.Time
	timestamps = append(timestamps, time.Now())
	for {
		r, ok := <-in.Req
		if !ok {
			break
		}

		switch r.Op {
		case cmn.GraphOp:
			g, _ := <-in.Graph
			fmt.Println("Graph", g)
		case cmn.KMST:
			<-in.Graph
			//fmt.Println("MST", g)
		case cmn.CurrTime:
			timestamps = append(timestamps, time.Now())
			fmt.Println("Ellapsed Time Since Last: ", timestamps[len(timestamps)-1].Sub(timestamps[len(timestamps)-2]))
		case cmn.EOF:
			<-in.Graph
			break
		default: //something's wrong
			fmt.Println("Unknown operation in output")
			break
		}
	}
	end <- struct{}{}
}

func generator(in in_comm, out out_comm, Fsize int, inputSync chan<- struct{}) {
	filter_count := 0
	for {
		r, ok := <-in.Req
		if !ok {
			break
		}
		switch r.Op {
		case cmn.Insert, cmn.Update:
			out_req := make(chan cmn.Request, channelSize)
			out_grph := make(chan Msg, channelSize)
			new_out := out_comm{Req: out_req, Graph: out_grph}
			new_in := in_comm{Req: out_req, Graph: out_grph}

			go filter(filter_count, in, new_out, r, Fsize)
			filter_count++
			in = new_in
		case cmn.Delete:
			// Do nothing, asked to delete unexistant edge
		case cmn.GraphOp, cmn.KMST:
			g, _ := <-in.Graph

			out.Req <- r
			out.Graph <- g
		case cmn.EOF:
			g, _ := <-in.Graph
			out.Req <- r
			out.Graph <- g
			out.Req <- r
			break
		case cmn.SaveState:
			// Do nothing
		case cmn.LoadState:
			// Generate the needed number of filters
			dir, err := os.ReadDir("./savestate")
			if err != nil {
				fmt.Println("Could not open folder savestate")
				break
			}
			if len(dir)%2 != 0 {
				fmt.Println("Savestate folder may be wrong")
				break
			}

			if filter_count != len(dir)/2 {
				out_req := make(chan cmn.Request, channelSize)
				out_grph := make(chan Msg, channelSize)
				new_out := out_comm{Req: out_req, Graph: out_grph}
				new_in := in_comm{Req: out_req, Graph: out_grph}

				go filter(filter_count, in, new_out, r, Fsize)
				filter_count++
				in = new_in
			} else {
				<-in.Graph
				inputSync <- struct{}{}
			}
		case cmn.CurrTime:
			// Do nothing
			out.Req <- r
		default: //something's wrong
			fmt.Println("Unknown operation in generator")
			break
		}
	}
	close(out.Req)
	close(out.Graph)
}

func filter(id int, in in_comm, out out_comm, r cmn.Request, Fsize int) {
	s := make(map[int64]struct{}, Fsize)

	int_comm := make(chan cmn.Request, 4096)

	go filter_worker(id, in_comm{int_comm, in.Graph}, out.Graph, Fsize)
	var ok bool

	for {
		switch r.Op {
		case cmn.Insert, cmn.Update:
			if _, ok = s[r.E.X]; ok {
				int_comm <- r
			} else if len(s) < Fsize {
				s[r.E.X] = struct{}{}
				int_comm <- r
			} else {
				out.Req <- r
			}
		case cmn.Delete:
			if _, ok = s[r.E.X]; ok {
				int_comm <- r
			} else {
				out.Req <- r
			}
		case cmn.KMST:
			int_comm <- r
			out.Req <- r
		case cmn.GraphOp:
			int_comm <- r
			out.Req <- r
		case cmn.SaveState:
			int_comm <- r
			out.Req <- r

			file, err := os.Create("./savestate/filter_" + strconv.Itoa(id) + ".bin")
			if err != nil {
				log.Fatal("Cannot open file for savestate of filer " + strconv.Itoa(id))
			}

			// Create a new gob encoder and use it to encode the person struct
			enc := gob.NewEncoder(file)
			if err := enc.Encode(s); err != nil {
				fmt.Println("Error encoding struct:", err)
				return
			}

			file.Close()
		case cmn.LoadState:
			int_comm <- r
			out.Req <- r

			file, err := os.Open("./savestate/filter_" + strconv.Itoa(id) + ".bin")
			if err != nil {
				log.Fatal("Cannot open savestate of filter " + strconv.Itoa(id))
			}

			dec := gob.NewDecoder(file)
			if err := dec.Decode(&s); err != nil {
				log.Fatal("Cannot load state of filter " + strconv.Itoa(id))
			}

			file.Close()

			fmt.Println("Filter ", id, "loaded state")
		case cmn.CurrTime:
			out.Req <- r
		case cmn.EOF:
			int_comm <- r
			out.Req <- r
			break
		}

		r, ok = <-in.Req
		if !ok {
			break
		}
	}

	close(out.Req)
}

func filter_worker(id int, in in_comm, out chan<- Msg, Fsize int) {

	// Initialize memory
	root := make(map[int64]*cmn.Graph, Fsize)
	self_update := false
	self_mst := make(cmn.Graph, 0)
	returnMST := make(cmn.Graph, 0, 50000)

	for {
		r, ok := <-in.Req
		if !ok {
			break
		}

		switch r.Op {
		case cmn.Insert, cmn.Update:
			if _, ok = root[r.E.X]; ok {
				root[r.E.X].InsertUpdate(r.E)
			} else {
				root[r.E.X] = &cmn.Graph{r.E}
			}
			self_update = true
		case cmn.Delete:
			if _, ok = root[r.E.X]; ok {
				root[r.E.X].Delete(r.E)
			}
			self_update = true
		case cmn.KMST:
			g, _ := <-in.Graph
			if g.Updated || self_update {
				g.Updated = mst.Kruskal(root, g.Graph, self_mst, &returnMST)
				self_mst = returnMST
				self_update = false
			}
			g.Graph = self_mst
			out <- g
		case cmn.GraphOp:
			g, _ := <-in.Graph
			for _, adje := range root {
				g.Graph = append(g.Graph, *adje...)
			}

			out <- g
		case cmn.SaveState:
			file, err := os.Create("savestate/worker_filter_" + strconv.Itoa(id) + ".bin")
			if err != nil {
				log.Fatal("Cannot open file for savestate of worker " + strconv.Itoa(id))
			}

			// Create a new gob encoder and use it to encode the person struct
			enc := gob.NewEncoder(file)
			if err := enc.Encode(root); err != nil {
				fmt.Println("Error encoding state of worker:", err)
				return
			}

			file.Close()

			g, _ := <-in.Graph
			out <- g
		case cmn.LoadState:
			g, _ := <-in.Graph

			file, err := os.Open("savestate/worker_filter_" + strconv.Itoa(id) + ".bin")
			if err != nil {
				log.Fatal("Cannot open savestate of worker " + strconv.Itoa(id))
			}

			dec := gob.NewDecoder(file)
			if err := dec.Decode(&root); err != nil {
				log.Fatal("Cannot load state of worker " + strconv.Itoa(id))
			}

			file.Close()

			fmt.Println("Worker ", id, "loaded state")
			out <- g
		case cmn.EOF:
			<-in.Graph
			out <- Msg{cmn.Graph{}, false}
			break
		}
	}

	close(out)
}

func Start(istream string, Fsize int) {

	file_req := make(chan cmn.Request, channelSize)
	file_grph := make(chan Msg, channelSize)
	file_gen := out_comm{Req: file_req, Graph: file_grph}
	gen := in_comm{Req: file_req, Graph: file_grph}

	gen_req := make(chan cmn.Request, channelSize)
	gen_grph := make(chan Msg, channelSize)
	gen_out := out_comm{Req: gen_req, Graph: gen_grph}
	out := in_comm{Req: gen_req, Graph: gen_grph}

	end := make(chan struct{})
	inputSync := make(chan struct{})

	start := time.Now()
	go input(istream, file_gen, Fsize, inputSync)
	go generator(gen, gen_out, Fsize, inputSync)
	go output(istream, out, end)
	<-end
	t := time.Since(start)
	fmt.Println("TotalExecutionTime,", t, ",", t.Microseconds(), "μs,", t.Milliseconds(), "ms ,", t.Seconds(), "s")
}

func RandomStart(Fsize int, Nmin, Nmax, Ndelta, Ngraph, Nrep int64, Dmin, Dmax, Ddelta float64) {
	f, _ := os.Create("RandomStart.csv")
	defer f.Close()
	f.WriteString("N;D;seed;microseconds\n")

	for N := Nmin; N < Nmax; N += Ndelta {
		for D := Dmin; D < Dmax; D += Ddelta {
			for i := Ngraph; i > 0; i-- {
				var seed int64 = rand.Int63()
				for j := Nrep; j > 0; j-- {
					input_req := make(chan cmn.Request, channelSize)
					input_grph := make(chan Msg, channelSize)
					input_chan := out_comm{Req: input_req, Graph: input_grph}
					data_chan := in_comm{Req: input_req, Graph: input_grph}

					gen_req := make(chan cmn.Request, channelSize)
					gen_grph := make(chan Msg, channelSize)
					gen_out := out_comm{Req: gen_req, Graph: gen_grph}
					out := in_comm{Req: gen_req, Graph: gen_grph}

					end := make(chan struct{})
					inputSync := make(chan struct{})

					start := time.Now()
					go ErdosRenyi(0.05, D, N, seed, input_chan, Fsize)
					go generator(data_chan, gen_out, Fsize, inputSync)
					go output("", out, end)
					<-end
					t := time.Since(start)

					fmt.Fprintf(f, "%d;%.4f;%d;%d\n", N, D, seed, t.Microseconds())
				}
			}
		}
	}
}
