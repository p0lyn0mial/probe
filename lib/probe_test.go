package probe_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	probe "github.com/probe/lib"
)

//TODO: add description
func TestStart(t *testing.T) {
	var scenarios = []struct {
		duration time.Duration
		rate     int
		target   string
		ctx      context.Context
		output   string
	}{
		// scenario 0: TODO: add description
		{
			duration: time.Duration(30) * time.Second,
			rate:     2,
			target:   "",
			ctx:      context.TODO(),
			output:   "Printing collected samples",
		},
	}

	for i, ts := range scenarios {
		buf := &bytes.Buffer{}
		target := probe.New()
		res := target.Start(ts.ctx, ts.duration, ts.rate, ts.target)
		err := res.Print(buf)
		if err != nil {
			t.Errorf("scenario %d: failed to print results, due to %s", i, err.Error())
		}
		if buf.String() != ts.output {
			t.Errorf("scenario %d: incorrect output returned\n Expected: %s\n Got: %s\n", i, buf.String(), ts.output)
		}
	}
}
