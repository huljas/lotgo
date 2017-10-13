package lotgo

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"runtime"
)

type Runner struct {
	clients       int
	runs          int
	duration      time.Duration
	sleep         time.Duration
	test          LoadTest
	done          bool
	errout        io.Writer
	rampup        time.Duration
	activeClients int32
	allListeners  Listener
	stopped       bool
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
	Duration time.Duration
	Start    time.Time
}

var _ EndCondition = &TimeCondition{}

func (c *TimeCondition) Run() bool {
	return time.Since(c.Start) < c.Duration
}

func (runner *Runner) Stop() {
	runner.stopped = true
}

func (runner *Runner) IsDone() bool {
	return runner.done
}

func (runner *Runner) setDone(done bool) {
	runner.done = done
}

func (runner *Runner) ActiveClients() int32 {
	return atomic.LoadInt32(&runner.activeClients)
}

func (runner *Runner) Run() {
	LOG().Infof("Runner starting run")
	LOG().Infof("**** clients=%d,runs=%d,time=%d,sleep=%d,GOMAXPROCS=%d,rampup=%d ****", runner.clients, runner.runs, runner.duration/1000/1000/1000, runner.sleep, runtime.GOMAXPROCS(0), runner.rampup)
	finishedWG := &sync.WaitGroup{}
	finishedWG.Add(runner.clients)
	setupWG := &sync.WaitGroup{}
	setupWG.Add(runner.clients)
	runner.allListeners.Started(runner)
	LOG().Infof("Listeners started")
	LOG().Infof("Runner starting clients")
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
	LOG().Infof("Clients started, waiting to finish")
	finishedWG.Wait()
	LOG().Infof("All clients done")
	runner.allListeners.Finished()
	LOG().Infof("Listeners finished")
	runner.setDone(true)
	LOG().Infof("Runner run complete")
}

func (runner *Runner) runTest(test LoadTest, setupWG *sync.WaitGroup) {
	end := runner.EndCondition()
	test.SetUp(runner)
	defer test.TearDown(runner)
	setupWG.Done()
	atomic.AddInt32(&runner.activeClients, 1)
	for end.Run() && !runner.stopped {
		ts := time.Now()
		err := test.Test(runner)
		dur := time.Since(ts)
		if err == nil {
			runner.allListeners.Success(dur)
		} else {
			runner.allListeners.Error(err)
		}
		if runner.sleep > 0 && end.Run() {
			time.Sleep(runner.sleep)
		}
	}
	atomic.AddInt32(&runner.activeClients, -1)
}

func (runner *Runner) logError(err error, dur time.Duration) {
	runner.errout.Write([]byte(fmt.Sprintf("%s (%dms) Error : %s\n", time.Now().Format(time.RFC822Z), dur/time.Millisecond, err.Error())))
}

func (runner *Runner) EndCondition() EndCondition {
	if runner.runs > 0 {
		return &CountCondition{count: runner.runs}
	} else {
		return &TimeCondition{Start: time.Now(), Duration: runner.duration}
	}
}

func (runner *Runner) NewErr(str string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(str, args))
}

func (runner *Runner) Fail(args ...interface{}) {
	if len(args) > 0 && args[0] != nil {
		LOG().Fatal(args)
	}
}
