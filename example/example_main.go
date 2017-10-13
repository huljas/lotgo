package main

import (
	"errors"
	"github.com/huljas/lotgo"
	"os"
	"strconv"
	"time"
)

type SleepTest struct {
	Sleep time.Duration
}

func (t *SleepTest) SetUp(tr *lotgo.Runner) {
	s := os.Getenv("EXAMPLE_SLEEP")
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
	count int32
}

func (t *ErrorTest) SetUp(tr *lotgo.Runner) {
}

func (t *ErrorTest) TearDown(tr *lotgo.Runner) {
}

func (t *ErrorTest) Test(tr *lotgo.Runner) error {
	t.count++
	if t.count % 10 == 0 {
		return errors.New("random error")
	}
	if t.count % 11 == 0 {
		return errors.New("weird error")
	}
	if t.count % 17 == 0 {
		return errors.New("system error")
	}
	return nil
}

var _ lotgo.LoadTest = &SleepTest{}
var _ lotgo.LoadTest = &ErrorTest{}

func main() {
	lotgo.AddTest("example/sleep", &SleepTest{Sleep: time.Second})
	lotgo.AddTest("example/error", &ErrorTest{})
	lotgo.Run()
}
