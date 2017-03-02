package probe_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
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
			duration: time.Duration(5) * time.Second,
			rate:     10,
			target:   "",
			ctx:      context.TODO(),
			output:   "Total = %s",
		},
	}

	// run test scenarios
	for i, ts := range scenarios {
		buf := &bytes.Buffer{}
		target, err := probe.New(ts.ctx, ts.duration, ts.rate, ts.target)
		if err != nil {
			t.Errorf("scenario %d: failed to create probe object, due to %s", i, err.Error())
		}

		// start probe
		start := time.Now()
		res := target.Start()
		elapsed := time.Since(start)

		// validate samples
		if res.Total <= 0 {
			t.Errorf("scenario %d: total number of request must be > 0", i)
		}

		err = res.Print(buf)
		ts.output = fmt.Sprintf(ts.output, res.Total)
		if err != nil {
			t.Errorf("scenario %d: failed to print results, due to %s", i, err.Error())
		}
		if strings.Contains(buf.String(), ts.output) {
			t.Errorf("scenario %d: incorrect output returned\n Got: %s\n ShouldContain: %s\n", i, buf.String(), ts.output)
		}

		// validate execution time with 1s margin
		var durationDiff time.Duration
		if elapsed > ts.duration {
			durationDiff = elapsed - ts.duration
		} else {
			durationDiff = ts.duration - elapsed
		}
		if durationDiff > time.Duration(1)*time.Second {
			t.Errorf("scenario %d: probe duration exeeded\n Elapsed %v, Wanted %v, Diff %v", i, elapsed, ts.duration, durationDiff)
		}
	}
}
