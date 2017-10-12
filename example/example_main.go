package main

import (
	"errors"
	"github.com/huljas/lotgo"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type SleepTest struct {
	Sleep time.Duration
}

func (t *SleepTest) SetUp(tr *lotgo.Runner) {
	s := os.Getenv("EXAMPLE_DELAY")
	if s != "" {
		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			t.Sleep = time.Duration(i) * time.Millisecond
		}
	}
}

func (t *SleepTest) TearDown(tr *lotgo.Runner) {
}

func (t *SleepTest) Test(tr *lotgo.Runner) error {
	time.Sleep(t.Sleep)
	return nil
}

type ErrorTest struct {
	Ratio int32
	count int32
}

func (t *ErrorTest) SetUp(tr *lotgo.Runner) {
}

func (t *ErrorTest) TearDown(tr *lotgo.Runner) {
}

func (t *ErrorTest) Test(tr *lotgo.Runner) error {
	i := atomic.AddInt32(&t.count, 1)
	if i % t.Ratio == 0 {
		return errors.New("Error")
	}
	return nil
}

var _ lotgo.LoadTest = &SleepTest{}
var _ lotgo.LoadTest = &ErrorTest{}

func main() {
	lotgo.AddTest("example/sleep", &SleepTest{Sleep: time.Millisecond})
	lotgo.AddTest("example/error", &ErrorTest{Ratio: 1})
	lotgo.Run()
}
