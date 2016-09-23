package main

/* Imports */
import (
	"errors"    // error handling
	"flag"      // command-line options parsing
	"fmt"       // Printf()
	"math/rand" // Rand()
	"net"       // net.Conn
	"os"        // os.Exit(), os.Signal, os.Stderr, ...
	"os/signal" // signal's handling
	"strconv"   // Atoi()
	"syscall"   // signal's handling
	"time"      // Sleep()
)

/* Constants */
const (
	charset       = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	charset_len   = len(charset)
	char_idx_bits = 7 // 7 bits for 128 character charset
	char_idx_mask = (1 << char_idx_bits) - 1
	char_idx_max  = 63 / char_idx_bits // number of letter indices fitting in 63 bits

	w_charset     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	w_charset_len = len(w_charset)
	w_sep         = " !(),-./:;<=>?" // word separators
        w_max_letters = 10

	PNAME         = "slstress"
	D_LEN         = 256     // default string length
	D_USECS       = 1000000 // default delay in microseconds
)

/* Global variables */
var msg_sent = 0 // messages sent to syslog

/* Options */
var p_seed = flag.Int("s", 0, "seed for Rand")
var p_string_length = flag.Int("l", D_LEN, "length of a string being sent through syslog")
var p_words = flag.Bool("w", false, "try to generate random \"words\" instead of random strings")
var p_quiet = flag.Bool("q", false, "do not print statistics")
var p_stderr = flag.Bool("e", false, "output the message to standard error as well as to the system log")
var usecs = D_USECS // sleep delay microseconds between syslog() calls
var tag = PNAME     // default tag

/* Functions */
func rand_string(n int) string {
	modulo := int32(charset_len - 1)
	b := make([]byte, n)

	for i := range b {
		b[i] = charset[rand.Int31()%modulo]
	}
	return string(b)
}

func rand_string_fast(n int) string {
	b := make([]byte, n)

	for i, cache, remain := n-1, rand.Int63(), char_idx_max; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), char_idx_max // generate 63 random bits, for `char_idx_max' letters
		}
		if idx := int(cache & char_idx_mask); idx < charset_len {
			b[i] = charset[idx]
			i--
		}
		cache >>= char_idx_bits
		remain--
	}

	return string(b)
}

func rand_words_fast(n int) string {
	b := make([]byte, n)

        i, cache, remain := n-1, rand.Int63(), char_idx_max
        word_boundary := (cache % w_max_letters) + 1
	for ; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), char_idx_max // generate 63 random bits, for `char_idx_max' letters
		}
		if idx := int(cache & char_idx_mask); idx < w_charset_len {
			if word_boundary == 0 {
				b[i] = ' '
			        word_boundary = (cache % w_max_letters) + 1
			} else {
				b[i] = w_charset[idx]
			}
			i--
	                word_boundary--
		}
		cache >>= char_idx_bits
		remain--
	}

	return string(b)
}

func unix_syslog() (conn net.Conn, e error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog", "/var/run/log"}

	for _, network := range logTypes {
		for _, path := range logPaths {
			conn, e := net.Dial(network, path)
			if e != nil {
				continue
			} else {
				return conn, nil
			}
		}
	}
	return nil, errors.New("Unix syslog delivery error")
}

func print_stats() {
	if(*p_quiet) { return }

	fmt.Fprintf(os.Stdout, "Messages sent: %d\n", msg_sent)
	fmt.Fprintf(os.Stdout, "String length: %d\n", *p_string_length)
	fmt.Fprintf(os.Stdout, "Delay (usecs): %d\n", usecs)
}

func sig_caught_usr1(sigChan chan os.Signal) {
	<-sigChan

	print_stats()
}

func sig_caught_exit(sigChan chan os.Signal) {
	<-sigChan

	print_stats()

	os.Exit(0)
}

func stats_panic(e error) {
	print_stats()
	panic(e)
}

func rand_perf() {
	for j := 0; j <= 10000000; j++ {
		rand_string(*p_string_length)
	}

	os.Exit(0)
}

func parse_cmd_opts() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <DELAY>\n", PNAME)
		fmt.Fprintf(os.Stderr, "Example: %s -t my_test -l 1024 %d\n\n", PNAME, usecs)
		fmt.Fprintf(os.Stderr, "Options:\n")

		flag.PrintDefaults()
	}
	flag.StringVar(&tag, "t", tag, "mark every line to be logged with specified tag")

	flag.Parse() // to execute the command-line parsing
}

func set_signals() {
	sig_chan := make(chan os.Signal)
	sig_chan_exit := make(chan os.Signal)

	signal.Notify(sig_chan, syscall.SIGUSR1)
	signal.Notify(sig_chan_exit, syscall.SIGINT, syscall.SIGTERM)

	go sig_caught_usr1(sig_chan)
	go sig_caught_exit(sig_chan_exit)
}

func syslog_spammer(string_length int, usecs int, tag string) {
	var rand_fn = rand_string_fast
	var log_to_stdout = false
	rand.Seed(int64(*p_seed))

	if(*p_words) {
		rand_fn = rand_words_fast
	}

	conn, e := unix_syslog()
	if e != nil {
		/* there was an error, log to stdout */
		log_to_stdout = true
		fmt.Fprintf(os.Stderr, "%s failed to connect to syslog, logging to stdout\n", PNAME)
	}
	for true {
		s := fmt.Sprintf("%s: %s", tag, rand_fn(*p_string_length))

		if (log_to_stdout) {
			fmt.Fprintf(os.Stdout, "%s\n", s)
		} else {
			_, e := conn.Write([]byte(s))
			if e != nil {
				stats_panic(e)
			}
		}
		if (*p_stderr) {
			/* also log to stderr option set */
			fmt.Fprintf(os.Stderr, "%s\n", s)
		}
		msg_sent++
		time.Sleep(time.Duration(usecs) * time.Microsecond)
	}
}

func main() {
	var e error

	set_signals()

	parse_cmd_opts()

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	usecs, e = strconv.Atoi(os.Args[len(os.Args)-1])
	if e != nil {
		fmt.Fprintf(os.Stderr, "<DELAY> `%s' not an integer\n", os.Args[1])
		os.Exit(1)
	}
//	fmt.Println(usecs)

	/* Workhorse */
	syslog_spammer(*p_string_length, usecs, tag)

	if(*p_quiet == false) { print_stats() }
}
