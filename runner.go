package lotgo

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Runner struct {
	clients       int
	runs          int
	duration      time.Duration
	sleep         time.Duration
	periodLogger  Logger
	summaryLogger Logger
	test          LoadTest
	done          bool
	errout        io.Writer
	rampup        time.Duration
	activeProcs   int32
}

type EndCondition interface {
	Run() bool
}

var _ EndCondition = &CountCondition{}

type CountCondition struct {
	count int
	i     int
}

func (c *CountCondition) Run() bool {
	c.i++
	return c.count >= c.i
}

type TimeCondition struct {
	End time.Time
}

var _ EndCondition = &TimeCondition{}

func (c *TimeCondition) Run() bool {
	return time.Now().Before(c.End)
}

func (runner *Runner) IsDone() bool {
	return runner.done
}

func (runner *Runner) SetDone(done bool) {
	runner.done = done
}

func (runner *Runner) Run() {
	log.Infof("Starting test...")
	log.Infof("**** clients=%d,runs=%d,time=%d,sleep=%d,GOMAXPROCS=%d,rampup=%d ****", runner.clients, runner.runs, runner.duration/1000/1000/1000, runner.sleep, runtime.GOMAXPROCS(0), runner.rampup)
	finishedWG := &sync.WaitGroup{}
	finishedWG.Add(runner.clients)
	setupWG := &sync.WaitGroup{}
	setupWG.Add(runner.clients)
	go runner.periodLogger.Log(runner)
	log.Infof("Test running...")
	for p := 0; p < runner.clients; p++ {
		myTest := deepClone(runner.test)
		go func() {
			if runner.rampup > 0 {
				time.Sleep(runner.rampup * time.Duration(p) / time.Duration(runner.clients))
			}
			runner.runTest(myTest, setupWG)
			finishedWG.Done()
		}()
	}
	finishedWG.Wait()
	runner.SetDone(true)
	runner.summaryLogger.Log(runner)
	log.Infof("Test done!")
}

func (runner *Runner) runTest(test LoadTest, setupWG *sync.WaitGroup) {
	end := runner.EndCondition()
	test.SetUp(runner)
	defer test.TearDown(runner)
	setupWG.Done()
	atomic.AddInt32(&runner.activeProcs, 1)
	for end.Run() {
		ts := time.Now()
		err := test.Test(runner)
		dur := time.Since(ts)
		if err == nil {
			runner.periodLogger.AddTime(dur)
			runner.summaryLogger.AddTime(dur)
		} else {
			runner.periodLogger.IncErr()
			runner.summaryLogger.IncErr()
			runner.logError(err, dur)
		}
		if runner.sleep > 0 && end.Run() {
			time.Sleep(runner.sleep)
		}
	}
	atomic.AddInt32(&runner.activeProcs, -1)
}

func (runner *Runner) logError(err error, dur time.Duration) {
	runner.errout.Write([]byte(fmt.Sprintf("%s (%dms) Error : %s\n", time.Now().Format(time.RFC822Z), dur/time.Millisecond, err.Error())))
}

func (runner *Runner) EndCondition() EndCondition {
	if runner.runs > 0 {
		return &CountCondition{count: runner.runs}
	} else {
		return &TimeCondition{End: time.Now().Add(runner.duration)}
	}
}

func (runner *Runner) NewErr(str string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(str, args))
}

func (runner *Runner) Fail(args ...interface{}) {
	if len(args) > 0 && args[0] != nil {
		log.Fatal(args)
	}
}
