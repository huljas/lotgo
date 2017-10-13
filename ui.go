package lotgo

import (
	"github.com/gizak/termui"
	"time"
	"github.com/rcrowley/go-metrics"
	"fmt"
	"runtime"
)

const (
	THROUGHPUT_COUNT = 72
)

type ui struct {
	runner       *Runner
	topText      *termui.Par
	testProgress *termui.Gauge
	summaryList  *termui.List
	errorList    *termui.List
	throughPut   *termui.LineChart

	totalTimer metrics.Timer
	lastUpdate time.Time
	lastCount  int64
	lastX      float64
	errors     metrics.Counter
	lastErrors []string
}

var _ Listener = &ui{}

func NewUi() *ui {
	return &ui{totalTimer: metrics.NewTimer(), errors: metrics.NewCounter()}
}

func (ui *ui) Loop() {
	if !terminalUi {
		return
	}
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	topText := termui.NewPar("PRESS q TO QUIT")
	topText.Height = 3
	topText.Width = 40
	topText.TextFgColor = termui.ColorWhite
	topText.BorderLabel = "Text Box"
	topText.BorderFg = termui.ColorCyan
	ui.topText = topText;

	testProgress := termui.NewGauge()
	testProgress.Percent = 50
	testProgress.Width = 40
	testProgress.Height = 3
	testProgress.Y = 0
	testProgress.BorderLabel = "Test progress"
	testProgress.BarColor = termui.ColorRed
	testProgress.BorderFg = termui.ColorWhite
	testProgress.BorderLabelFg = termui.ColorCyan
	testProgress.X = 41
	ui.testProgress = testProgress

	strs := []string{"[0] gizak/termui", "[1] editbox.go", "[2] interrupt.go", "[3] keyboard.go", "[4] output.go", "[5] random_out.go", "[6] dashboard.go", "[7] nsf/termbox-go"}
	errorList := termui.NewList()
	errorList.Items = strs
	errorList.ItemFgColor = termui.ColorYellow
	errorList.BorderLabel = "Errors"
	errorList.Height = 13
	errorList.Width = 40
	errorList.Y = 3
	errorList.X = 41
	ui.errorList = errorList

	strs2 := []string{"Throughput: 100.3 r/s", "Response time, avg: 10 ms", "      95%: 128 ms", "      75%: 45 ms", "Count: 12312", "Active clients: 10"}
	summaryList := termui.NewList()
	summaryList.Items = strs2
	summaryList.ItemFgColor = termui.ColorWhite
	summaryList.BorderLabel = "Summary"
	summaryList.Height = 13
	summaryList.Width = 40
	summaryList.Y = 3
	summaryList.X = 0
	ui.summaryList = summaryList

	throughPut := termui.NewLineChart()
	throughPut.BorderLabel = "Throughput r/s"
	throughPut.Width = 81
	throughPut.Height = 11
	throughPut.X = 0
	throughPut.Y = 16
	throughPut.AxesColor = termui.ColorWhite
	throughPut.LineColor = termui.ColorRed | termui.AttrBold
	throughPut.Mode = "dot"
	throughPut.Data = []float64{0.0}
	ui.throughPut = throughPut

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
		ui.runner.Stop()
	})
	termui.Handle("/timer/1s", func(e termui.Event) {
		t := e.Data.(termui.EvtTimer)
		ui.Redraw(int(t.Count))
	})

	termui.Loop()
	LOG().Infof("termui closed")
}

func (ui *ui) Started(runner *Runner) {
	ui.runner = runner
}

func (ui *ui) Finished() {
	LOG().Infof("Stopping UI on test finished")
	termui.StopLoop()
}

func (ui *ui) Success(d time.Duration) {
	ui.totalTimer.Update(d)
}

func (ui *ui) Error(err error) {
	list := ui.lastErrors
	list = append(list, err.Error())
	if len(list) > 10 {
		list = list[len(list)-10:]
	}
	ui.lastErrors = list
	ui.errors.Inc(1)
}

func (ui *ui) Redraw(int) {
	if ui.testProgress == nil {
		return
	}
	ui.testProgress.Percent = int(ui.calculateProgress())
	ui.errorList.Items = ui.lastErrors
	ui.summaryList.Items = ui.summaryItems()

	xData := ui.throughPut.Data
	xData = append(xData, ui.lastX)
	if len(xData) > THROUGHPUT_COUNT {
		xData = xData[len(xData) - THROUGHPUT_COUNT:]
	}
	ui.throughPut.Data = xData

	termui.Render(ui.topText, ui.testProgress, ui.summaryList, ui.errorList, ui.throughPut)
}

func (ui *ui) calculateProgress() int64 {
	if ui.runner.runs > 0 {
		total := ui.runner.runs * ui.runner.clients
		count := ui.totalTimer.Count()
		return count * 100 / int64(total)
	} else {
		duration := ui.runner.duration
		since := time.Since(ui.runner.startTime)
		return int64(since * 100 / duration)
	}
}

func (ui *ui) summaryItems() []string {
	totalCount := ui.totalTimer.Count()
	if !ui.lastUpdate.IsZero() {
		count := totalCount - ui.lastCount
		duration := time.Since(ui.lastUpdate)
		ui.lastX = float64(count) / float64(duration) * float64(time.Second)
	}
	ui.lastUpdate = time.Now()
	ui.lastCount = totalCount
	count := ui.totalTimer.Count()
	since := time.Since(ui.runner.startTime)
	return []string{
		fmt.Sprintf("Test:                %s", testName),
		fmt.Sprintf("Go max procs:        %d", runtime.GOMAXPROCS(0)),
		fmt.Sprintf("Clients:             %d / %d", ui.runner.ActiveClients(), clients),
		fmt.Sprintf("Errors:              %d", ui.errors.Count()),
		fmt.Sprintf("Successes:           %d", count),
		fmt.Sprintf("Time:                %d ms", since / time.Millisecond),
		fmt.Sprintf("Throughput:          %f r/s", ui.lastX),
		fmt.Sprintf("Response time, mean: %f ms", ui.totalTimer.Mean()/1000/1000),
		fmt.Sprintf("Response time, 75%%:  %f ms", ui.totalTimer.Percentile(0.75)/1000/1000),
		fmt.Sprintf("Response time, 95%%:  %f ms", ui.totalTimer.Percentile(0.95)/1000/1000),
	}
}