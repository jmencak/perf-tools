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
	"time"      // Sleep()
)

/* Constants */
const (
	PNAME         = "gotime"
	DOC_ROOT      = "."
)

/* Global variables */
var requests = 0

/* Options */
var p_verbose = flag.Bool("v", false, "verbose mode")
var p_sleep = flag.Int("s", 5, "sleep between checks")
var p_req_quit = flag.Int("n", 0, "quit after receiving #n GOTIME requests (0: never quit)")
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

func http_srv_gotime(w http.ResponseWriter, req *http.Request) {
	if *p_verbose {
		log.Printf("Received %d. request %q\n", requests, req.URL.Path)
	}

	requests++
	responseString := "GOTIME"
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(responseString)))
	io.WriteString(w, responseString)

	if *p_req_quit != 0 && requests >= *p_req_quit {
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

	if *p_verbose {
		log.Printf("Received %d. request %q\n", requests, req.URL.Path)
	}

	data, err := ioutil.ReadFile(p_doc_root + req.URL.Path)
	if err != nil {
		log.Printf("error reading %v: %v", req.URL.Path, err)
		status = http.StatusNotFound	// 404
		data = []byte("404 Not Found")
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/GOTIME", http_srv_gotime)
	mux.HandleFunc("/", http_srv_file)	// catch-all to serve files/configuration

	parse_cmd_opts()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

//	fmt.Fprintf(os.Stdout, "cmd=%s, arguments=%s\n", flag.Args()[0], flag.Args()[1:])

	for true {
		_, err := run_cmd_args(flag.Args()[0],flag.Args()[1:])
//		if *p_verbose {
//			log.Printf("stdout: %s\n", cmd_out)
//		}
		if err == nil {
			// command succeeded (return status 0)
			break
		}
		// command failed
		if *p_verbose {
			log.Printf("`%s' failed, sleeping %d\n", flag.Args()[0], *p_sleep)
		}
		time.Sleep(time.Duration(*p_sleep) * time.Second)
	}

	log.Printf("Listening on port %d\n", *p_port)
	if *p_req_quit != 0 {
		log.Printf("Blocking until receiving %d GOTIME requests.\n", *p_req_quit)
	} else {
		log.Printf("Blocking forever.\n")
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf((":%d"),*p_port), mux))
}
