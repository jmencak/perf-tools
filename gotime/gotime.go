package main

import (
	"strconv"   // strconv
	"log"	    // log.Fatal()
	"io"	    // io.WriteString()
	"io/ioutil" // ioutil.ReadFile()
	"flag"      // command-line options parsing
	"net/http"  // http server
	"os"        // os.Exit(), os.Signal, os.Stderr, ...
	"fmt"       // Printf()
	"os/exec"   // exec.Command()
)

/* Constants */
const (
	PNAME         = "gotime"
	DOC_ROOT      = "."
)
const ( // states
	S_INIT        = 0
	S_RACE        = 1
	S_FINISHED    = 2
)

/* Global variables */
var clients_running = 0		// number of clients running (even after some have finished)
var clients_finished = 0	// number of clients that finished
var state = S_INIT

/* Options */
var p_verbose = flag.Bool("v", false, "verbose mode")
var p_req_quit = flag.Int("n", 0, "quit after receiving #n gotime/finish requests (0: never quit)")
var p_port = flag.Int("p", 9090, "port to listen on")

var p_doc_root = DOC_ROOT // document root

func parse_cmd_opts() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <COMMAND> [<ARGS>...]\n", PNAME)
		fmt.Fprintf(os.Stderr, "Example: %s -r /var/www/html -v -- file -f /tmp/gotime\n\n", PNAME)
		fmt.Fprintf(os.Stderr, "Options:\n")

		flag.PrintDefaults()
	}
	flag.StringVar(&p_doc_root, "r", p_doc_root, "path to gotime document root")
	flag.Parse() // to execute the command-line parsing
}

func run_cmd(cmd string) (string, error) {
	command := exec.Command("bash", "-c", cmd)
	out, err := command.Output()
	return string(out), err
}

func run_cmd_args(cmd string, args []string) (string, error) {
	command := exec.Command(cmd, args...)
	out, err := command.Output()
	return string(out), err
}

func http_srv_gotime_start(w http.ResponseWriter, req *http.Request) {
	var responseString = "GO"

	if state == S_INIT {
		/* haven't given gotime so far */
		_, err := run_cmd_args(flag.Args()[0],flag.Args()[1:])
//		if *p_verbose {
//			log.Printf("stdout: %s\n", cmd_out)
//		}
		if err == nil {
			// command succeeded (return status 0)
			state = S_RACE
		} else {
			// command failed
			responseString = "NOGO"
			if *p_verbose {
				log.Printf("`%s' failed, not ready to give a go\n", flag.Args()[0])
			}
		}
	}

	if state == S_RACE {
		/* race is on */
		if *p_req_quit != 0 && clients_running == *p_req_quit {
			/* already received *p_req_quit requests, what is this extra client! */
			responseString = "NOGO"
			if *p_verbose {
				log.Printf("Extra client detected, already received %d requests to start, sending NOGO\n", *p_req_quit)
			}
		} else {
			clients_running++
			if *p_verbose {
				log.Printf("Received %d. request %q\n", clients_running, req.URL.Path)
			}
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(responseString)))
	io.WriteString(w, responseString)
}

func http_srv_gotime_finish(w http.ResponseWriter, req *http.Request) {
	var responseString = "OK"

	if clients_running < (clients_finished + 1) {
		/* haven't given a go ahead so far, ignore this request */
		responseString = "KO"
		if *p_verbose {
			log.Printf("Out of order request for finish, %d/%d (running/finished)\n", clients_running, clients_finished)
		}
	} else {
		/* a valid request from a client that finished */
		clients_finished++
		if *p_verbose {
			log.Printf("%d/%d clients finished\n", clients_finished, *p_req_quit)
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(responseString)))
	io.WriteString(w, responseString)

	if *p_req_quit != 0 && clients_finished == *p_req_quit && state == S_RACE {
		state = S_FINISHED
		f, canFlush := w.(http.Flusher)
		if canFlush {
			f.Flush()
		}

		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			log.Fatalf("error while shutting down: %v", err)
		}

		conn.Close()

		log.Println("Shutting down")
		os.Exit(0)
	}
}

func http_srv_file(w http.ResponseWriter, req *http.Request) {
	status := http.StatusOK

	if req.URL.Path == "/" {
		// for compatibility with the original WLG
		http_srv_gotime_start(w, req)
		return
	}

	if *p_verbose {
		log.Printf("Received request %q\n", req.URL.Path)
	}

	data, err := ioutil.ReadFile(p_doc_root + req.URL.Path)
	if err != nil {
		log.Printf("error reading %v: %v", req.URL.Path, err)
		status = http.StatusNotFound	// 404
		data = []byte("")		// return empty string for default values
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func main() {
	parse_cmd_opts()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/gotime/start", http_srv_gotime_start)
	mux.HandleFunc("/gotime/finish", http_srv_gotime_finish)
	mux.HandleFunc("/", http_srv_file)	// catch-all to serve files/configuration

//	fmt.Fprintf(os.Stdout, "cmd=%s, arguments=%s\n", flag.Args()[0], flag.Args()[1:])

	log.Printf("Listening on port %d\n", *p_port)
	if *p_req_quit != 0 {
		log.Printf("Blocking until receiving %d gotime/finish requests.\n", *p_req_quit)
	} else {
		log.Printf("Blocking forever.\n")
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf((":%d"),*p_port), mux))
}
