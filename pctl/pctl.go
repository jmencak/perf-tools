package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

var numbers sort.Float64Slice

func main() {
	err := percentile(os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		panic(err)
	}
}

func percentile(r io.Reader, stdout io.Writer, stderr io.Writer) error {
	reader := bufio.NewReader(r)
	var line []byte
	var err error
	for {
		if line, _, err = reader.ReadLine(); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		f, convErr := strconv.ParseFloat(string(line), 64)
		if convErr != nil {
			fmt.Fprintf(stderr, "number conversion error: %s\n", convErr)
			continue
		}

		numbers = append(numbers, f)
	}

	sort.Sort(numbers)
	l := len(numbers)

//	for i := 1; i <= 100; i++ {
//		printPercentileN(stdout, &numbers, l, i, true)
//	}
	printPercentileN(stdout, &numbers, l, 90, false)
	printPercentileN(stdout, &numbers, l, 95, true)
	printPercentileN(stdout, &numbers, l, 99, true)
	fmt.Fprintf(stdout, "\n")

	return nil
}

func percentileN(numbers *sort.Float64Slice, l, n int) float64 {
	i := l*(n-1)/100
	ns := *numbers

	return ns[i]
}

func printPercentileN(w io.Writer, numbers *sort.Float64Slice, l, n int, comma bool) {
        if comma { fmt.Fprintf(w, "\t") }
	fmt.Fprintf(w, "%s", strconv.FormatFloat(percentileN(numbers, l, n), 'g', 16, 64))
}
