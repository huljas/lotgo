package lotgo

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

var _ LoadTest = &myTest{}

type myTest struct {
	Count int
	Value int
}

func (e *myTest) SetUp(lt *Runner) {
}

func (e *myTest) TearDown(lt *Runner) {
}

var myCount int32

func (e *myTest) Test(lt *Runner) error {
	atomic.AddInt32(&myCount, 1)
	return nil
}

func TestRunner_NumberOfTestRunsSingleProc(t *testing.T) {
	myCount = 0
	example := &myTest{}
	runner := New(1, 1000, 0, 0, time.Second, example, nil, nil, 0)
	runner.Run()
	assert.Equal(t, int32(1000), myCount)
}

func TestRunner_RightNumberOfTestRunsMultiProc(t *testing.T) {
	myCount = 0
	example := &myTest{}
	runner := New(20, 500, 0, 0, time.Second, example, nil, nil, 0)
	runner.Run()
	assert.Equal(t, int32(10000), myCount)
}

func TestRunner_TimeDuration(t *testing.T) {
	oneSec := time.Now().Add(time.Second)
	e := &TimeCondition{End: oneSec}
	assert.True(t, e.Run())
	time.Sleep(time.Millisecond * 500)
	assert.True(t, e.Run())
	time.Sleep(time.Millisecond * 500)
	assert.False(t, e.Run())
}

func TestRunner_RightDurationOfTestRun(t *testing.T) {
	ts := time.Now()
	example := &myTest{}
	lt := New(20, 0, time.Millisecond*100, 0, time.Second, example, nil, nil, 0)
	lt.Run()
	d := time.Since(ts)
	assert.True(t, time.Millisecond*100 < d)
	assert.True(t, time.Millisecond*200 > d)
}
