package lotgo

import (
	"fmt"
	"github.com/rcrowley/go-metrics"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// Logger collects data from the test
type Logger interface {
	Log(lt *Runner)
	AddTime(time time.Duration)
	IncErr()
}

// Format is used for formatting the results
type Format interface {
	FormatHeader() []string
	Format(res *Result) string
}

var _ Logger = &NoOpLogger{}

// NoOpLogger does nothing
type NoOpLogger struct {
}

/* No op */
func (l *NoOpLogger) AddTime(d time.Duration) {
}

/* No op */
func (l *NoOpLogger) IncErr() {
}

/* No op */
func (l *NoOpLogger) Log(lt *Runner) {
}

var _ Logger = &PeriodLogger{}

/* Logger which writes results periodically */
type PeriodLogger struct {
	sync.Mutex
	period      time.Duration
	start       time.Time
	timer       metrics.Timer
	errors      metrics.Counter
	format      Format
	writer      io.Writer
	periodStart time.Time
}

/* Returns new period logger */
func NewPeriodLogger(p time.Duration, w io.Writer, f Format) *PeriodLogger {
	now := time.Now()
	l := &PeriodLogger{period: p, start: now, writer: w, format: f}
	l.newPeriod()
	return l
}

func (l *PeriodLogger) AddTime(d time.Duration) {
	l.Lock()
	l.timer.Update(d)
	l.Unlock()
}

func (l *PeriodLogger) IncErr() {
	l.Lock()
	l.errors.Inc(1)
	l.Unlock()
}

func (l *PeriodLogger) Log(lt *Runner) {
	l.printHead()
	for !lt.IsDone() {
		time.Sleep(l.period)
		if !lt.IsDone() {
			l.print(lt)
			l.newPeriod()
		}
	}
}

func (l *PeriodLogger) newPeriod() {
	l.Lock()
	l.periodStart = time.Now()
	l.timer = metrics.NewTimer()
	l.errors = metrics.NewCounter()
	l.Unlock()
}

func (l *PeriodLogger) printHead() {
	for _, s := range l.format.FormatHeader() {
		l.writer.Write([]byte(s))
	}
}

func (l *PeriodLogger) print(lt *Runner) {
	l.Lock()
	res := NewResult(l.start, l.periodStart, l.timer, l.errors, atomic.LoadInt32(&lt.activeProcs))
	line := l.format.Format(res)
	l.writer.Write([]byte(line))
	l.Unlock()
}

type SummaryLogger struct {
	warmup time.Duration
	start  time.Time
	timer  metrics.Timer
	errors metrics.Counter
	format Format
	writer io.Writer
	active bool
}

func NewSummaryLogger(warmup time.Duration, w io.Writer, f Format) *SummaryLogger {
	l := &SummaryLogger{warmup: warmup, start: time.Now(), format: f, writer: w}
	return l
}

var _ Logger = &SummaryLogger{}

func (l *SummaryLogger) startLog() {
	l.start = time.Now()
	l.timer = metrics.NewTimer()
	l.errors = metrics.NewCounter()
	l.active = true
}

func (l *SummaryLogger) IsActive() bool {
	if !l.active && time.Since(l.start) > l.warmup {
		l.startLog()
	}
	return l.active
}

func (l *SummaryLogger) AddTime(d time.Duration) {
	if l.IsActive() {
		l.timer.Update(d)
	}
}

func (l *SummaryLogger) IncErr() {
	if l.IsActive() {
		l.errors.Inc(1)
	}
}

func (l *SummaryLogger) Log(lt *Runner) {
	l.printHead()
	l.print(lt)
}

func (l *SummaryLogger) printHead() {
	for _, s := range l.format.FormatHeader() {
		l.writer.Write([]byte(s))
	}
}

func (l *SummaryLogger) print(lt *Runner) {
	res := NewResult(l.start, l.start, l.timer, l.errors, atomic.LoadInt32(&lt.activeProcs))
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

func (f *MdFormat) Format(res *Result) string {
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

func (f *CsvFormat) Format(res *Result) string {
	return fmt.Sprintf("%d,%d,%.1f,%.1f,%.1f,%.2f,%d\n", int64(res.Time.Seconds()), res.Count, res.Mean, res.P75, res.P95, res.Rate, res.ErrCount)
}
