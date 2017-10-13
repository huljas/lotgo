package lotgo

import (
	"fmt"
	"github.com/rcrowley/go-metrics"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// Format is used for formatting the results
type Format interface {
	FormatHeader() []string
	Format(res *result) string
}

// result is a summary of test results
type result struct {
	Time          time.Duration
	Count         int64
	Mean          float64
	P75           float64
	P95           float64
	Rate          float64
	ErrCount      int64
	ActiveClients int32
}

// NewResult creates a new result
func NewResult(start time.Time, periodStart time.Time, timer metrics.Timer, errCounter metrics.Counter, activeClients int32) *result {
	t := time.Since(start)
	p := time.Since(periodStart)
	c := timer.Count()
	m := timer.Mean() / 1000 / 1000
	p75 := timer.Percentile(0.75) / 1000 / 1000
	p95 := timer.Percentile(0.95) / 1000 / 1000
	rate := float64(c) / p.Seconds()
	ec := errCounter.Count()
	return &result{Time: t, Count: c, Mean: m, P75: p75, P95: p95, Rate: rate, ErrCount: ec, ActiveClients: activeClients}
}

var _ Listener = &periodLogger{}

/* Logger which writes results periodically */
type periodLogger struct {
	sync.Mutex
	period      time.Duration
	start       time.Time
	timer       metrics.Timer
	errors      metrics.Counter
	format      Format
	writer      io.Writer
	periodStart time.Time
	running     bool
	runner 		*Runner
}

/* Returns new period logger */
func NewPeriodLogger(p time.Duration, w io.Writer, f Format) *periodLogger {
	now := time.Now()
	l := &periodLogger{period: p, start: now, writer: w, format: f}
	l.newPeriod()
	return l
}

func (l *periodLogger) Success(d time.Duration) {
	l.Lock()
	l.timer.Update(d)
	l.Unlock()
}

func (l *periodLogger) Error(error) {
	l.Lock()
	l.errors.Inc(1)
	l.Unlock()
}

func (l *periodLogger) Started(runner *Runner) {
	l.running = true
	l.runner = runner
	go l.run()
}

func (l *periodLogger) Finished() {
	l.running = false
}

func (l *periodLogger) run() {
	first := true
	for l.running {
		time.Sleep(l.period)
		if l.running {
			if first {
				l.printHead()
				first = false
			}
			l.print()
			l.newPeriod()
		}
	}
}

func (l *periodLogger) newPeriod() {
	l.Lock()
	l.periodStart = time.Now()
	l.timer = metrics.NewTimer()
	l.errors = metrics.NewCounter()
	l.Unlock()
}

func (l *periodLogger) printHead() {
	for _, s := range l.format.FormatHeader() {
		l.writer.Write([]byte(s))
	}
}

func (l *periodLogger) print() {
	l.Lock()
	res := NewResult(l.start, l.periodStart, l.timer, l.errors, l.runner.ActiveClients())
	line := l.format.Format(res)
	l.writer.Write([]byte(line))
	l.Unlock()
}

type summaryLogger struct {
	warmup time.Duration
	start  time.Time
	timer  metrics.Timer
	errors metrics.Counter
	format Format
	writer io.Writer
	active bool
	runner *Runner
}

func NewSummaryLogger(warmup time.Duration, w io.Writer, f Format) *summaryLogger {
	l := &summaryLogger{warmup: warmup, start: time.Now(), format: f, writer: w, timer: metrics.NewTimer(), errors: metrics.NewCounter()}
	return l
}

var _ Listener = &summaryLogger{}

func (l *summaryLogger) startLog() {
	l.start = time.Now()
	l.active = true
}

func (l *summaryLogger) checkActive() bool {
	if !l.active && time.Since(l.start) > l.warmup {
		l.startLog()
	}
	return l.active
}

func (l *summaryLogger) Success(d time.Duration) {
	if l.checkActive() {
		l.timer.Update(d)
	}
}

func (l *summaryLogger) Error(error) {
	if l.checkActive() {
		l.errors.Inc(1)
	}
}

func (l *summaryLogger) Started(runner *Runner) {
	l.runner = runner
}

func (l *summaryLogger) Finished() {
	l.printHead()
	l.print(l.runner)
}

func (l *summaryLogger) printHead() {
	for _, s := range l.format.FormatHeader() {
		l.writer.Write([]byte(s))
	}
}

func (l *summaryLogger) print(lt *Runner) {
	res := NewResult(l.start, l.start, l.timer, l.errors, atomic.LoadInt32(&lt.activeClients))
	line := l.format.Format(res)
	l.writer.Write([]byte(line))
}

type MdFormat struct {
}

func (f *MdFormat) FormatHeader() []string {
	return []string{
		"| time     | count    | mean      | p75       | p95       | rate       | errs     | clients  |\n",
		"| -------- | -------- | --------- | --------- | --------- | ---------- | -------- | -------- |\n"}
}

func (f *MdFormat) getFormatString(delim string) string {
	return delim + " %8d " + delim + " %8d " + delim + " %9.1f " + delim + " %9.1f " + delim + " %9.1f " + delim + " %10.2f " + delim + " %8d " + delim + " %8d " + delim + "\n"
}

func (f *MdFormat) Format(res *result) string {
	delim := "|"
	fmtString := f.getFormatString(delim)
	return fmt.Sprintf(fmtString, int64(res.Time.Seconds()), res.Count, res.Mean, res.P75, res.P95, res.Rate, res.ErrCount, res.ActiveClients)
}

var _ Format = &MdFormat{}

type CsvFormat struct {
}

var _ Format = &CsvFormat{}

func (f *CsvFormat) FormatHeader() []string {
	return []string{"time,count,mean,p75,p95,rate,errs\n"}
}

func (f *CsvFormat) Format(res *result) string {
	return fmt.Sprintf("%d,%d,%.1f,%.1f,%.1f,%.2f,%d\n", int64(res.Time.Seconds()), res.Count, res.Mean, res.P75, res.P95, res.Rate, res.ErrCount)
}
