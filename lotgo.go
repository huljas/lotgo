package lotgo

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// NewFromCommandline creates new runner using commandline arguments
func NewFromCommandline() *Runner {
	var clients int
	var runs int
	var testName string
	var duration time.Duration
	var period time.Duration
	var sleep time.Duration
	var summaryFile string
	var errorLog string
	var maxprocs int
	var rampup time.Duration

	flag.IntVar(&clients, "clients", 1, "Number of clients to simulate")
	flag.IntVar(&runs, "runs", 1, "Number of runs per client")
	flag.StringVar(&testName, "test", "", "Name of the test, required. Allowed values: "+AllTests())
	flag.DurationVar(&duration, "duration", 0, "Duration of the test, overrides runs")
	flag.DurationVar(&period, "period", time.Second*10, "Period for logging the results")
	flag.StringVar(&summaryFile, "summaryFile", "", "Csv summary file, default stdout")
	flag.StringVar(&errorLog, "error", "", "Error log file, default stdout")
	flag.DurationVar(&sleep, "sleep", 0, "Time to sleep between test calls")
	flag.IntVar(&maxprocs, "maxprocs", 10, "Maximum number of goprocs")
	flag.DurationVar(&rampup, "rampup", 0, "Time to rampup all clients running")
	flag.Parse()

	runtime.GOMAXPROCS(maxprocs)

	if testName == "" {
		fmt.Println("-test is required")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if duration > 0 {
		runs = 0
	}
	test, ok := GetTest(testName)
	if !ok {
		fmt.Println("-test is unknown")
		flag.PrintDefaults()
		os.Exit(1)
	}
	var sw io.Writer
	if summaryFile != "" {
		f, err := os.Create(summaryFile)
		if err != nil {
			fmt.Printf("Failed to open '%s', reason %v", summaryFile, err)
		}
		fmt.Printf("Writing summary to '%s'\n", summaryFile)
		sw = f
	} else {
		sw = os.Stdout
	}

	var ew io.Writer
	if errorLog != "" {
		f, err := os.Create(errorLog)
		if err != nil {
			fmt.Printf("Failed to open '%s', reason %v", errorLog, err)
		}
		fmt.Printf("Writing errors to '%s'\n", errorLog)
		ew = f
	} else {
		ew = os.Stderr
	}
	return New(clients, runs, duration, sleep, period, test, sw, ew, rampup)
}

// New creates new runner
func New(clients int, runs int, duration time.Duration, sleep time.Duration, period time.Duration, test LoadTest, sw io.Writer, errw io.Writer, rampup time.Duration) *Runner {
	if period <= 0 {
		panic("Log period is less than 0!")
	}
	plogger := NewPeriodLogger(period, os.Stdout, &MdFormat{})
	var slogger Logger
	if sw != nil {
		slogger = NewSummaryLogger(duration/5, sw, &CsvFormat{})
	} else {
		slogger = NewSummaryLogger(duration/5, os.Stdout, &CsvFormat{})
	}

	return &Runner{clients: clients, runs: runs, duration: duration, sleep: sleep, test: test, periodLogger: plogger, summaryLogger: slogger, errout: errw, rampup: rampup}
}

// Run runs the test based on the commandline arguments
func Run() {
	runner := NewFromCommandline()
	runner.Run()
	os.Exit(0)
}
