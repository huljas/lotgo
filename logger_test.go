package lotgo_test

import (
	"bytes"
	"github.com/huljas/lotgo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMdFormat(t *testing.T) {
	f := lotgo.MdFormat{}
	expectedHeader := []string{
		"| time     | count    | mean      | p75       | p95       | rate       | errs     | clients  |\n",
		"| -------- | -------- | --------- | --------- | --------- | ---------- | -------- | -------- |\n"}
	assert.Equal(t, expectedHeader, f.FormatHeader())
	res := &lotgo.Result{Time: time.Second * 120, Count: 1000, Mean: 123.453243, P75: 134.3212412312, P95: 199.022311, Rate: 100.32332, ErrCount: 10, ActiveClients: 19}
	assert.Equal(t, "|      120 |     1000 |     123.5 |     134.3 |     199.0 |     100.32 |       10 |       19 |\n", f.Format(res))
	res = &lotgo.Result{Time: time.Second * 120, Count: 1000, Mean: 120000.453243, P75: 134000.3212412312, P95: 199000.022311, Rate: 100000.32332, ErrCount: 10, ActiveClients: 2}
	assert.Equal(t, "|      120 |     1000 |  120000.5 |  134000.3 |  199000.0 |  100000.32 |       10 |        2 |\n", f.Format(res))
	res = &lotgo.Result{Time: 0, Count: 0, Mean: 0, P75: 0, P95: 0.00, Rate: 0, ErrCount: 0}
	assert.Equal(t, "|        0 |        0 |       0.0 |       0.0 |       0.0 |       0.00 |        0 |        0 |\n", f.Format(res))
}

func TestCsvFormat(t *testing.T) {
	f := lotgo.CsvFormat{}
	assert.Equal(t, []string{"time,count,mean,p75,p95,rate,errs\n"}, f.FormatHeader())
	res := &lotgo.Result{Time: time.Second * 120, Count: 1000, Mean: 123.453243, P75: 134.3212412312, P95: 199.022311, Rate: 100.32332, ErrCount: 10}
	assert.Equal(t, "120,1000,123.5,134.3,199.0,100.32,10\n", f.Format(res))
}

func TestSummaryLoggerGracePeriod(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))
	l := lotgo.NewSummaryLogger(time.Millisecond*100, buf, &lotgo.CsvFormat{})
	assert.False(t, l.IsActive())
	time.Sleep(time.Millisecond * 100)
	assert.True(t, l.IsActive())
}
