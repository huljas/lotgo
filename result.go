package lotgo

import (
	"github.com/rcrowley/go-metrics"
	"time"
)

// Result is a summary of test results
type Result struct {
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
func NewResult(start time.Time, periodStart time.Time, timer metrics.Timer, errCounter metrics.Counter, activeClients int32) *Result {
	t := time.Since(start)
	p := time.Since(periodStart)
	c := timer.Count()
	m := timer.Mean() / 1000 / 1000
	p75 := timer.Percentile(0.75) / 1000 / 1000
	p95 := timer.Percentile(0.95) / 1000 / 1000
	rate := float64(c) / p.Seconds()
	ec := errCounter.Count()
	return &Result{Time: t, Count: c, Mean: m, P75: p75, P95: p95, Rate: rate, ErrCount: ec, ActiveClients: activeClients}
}
