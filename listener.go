package lotgo

import "time"

type Listener interface {
	Started(runner *Runner)
	Success(time time.Duration)
	Error(err error)
	Finished()
}

type Listeners []Listener

func (c Listeners) Add(l Listener) Listeners {
	return append(c, l)
}

func (c Listeners) Started(runner *Runner) {
	for _, l := range c {
		l.Started(runner)
	}
}

func (c Listeners) Success(d time.Duration) {
	for _, l := range c {
		l.Success(d)
	}
}

func (c Listeners) Error(err error) {
	for _, l := range c {
		l.Error(err)
	}
}

func (c Listeners) Finished() {
	for _, l := range c {
		l.Finished()
	}
}