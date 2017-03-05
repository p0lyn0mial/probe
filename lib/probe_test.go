package probe_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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
		ctxFunc  func() context.Context
		output   string
		// if > 0 used instead of duraiton
		ctxDuration time.Duration
		// test server calls handler before sending a response to a client.
		handler func()
	}{
		// scenario 0: happy path scenario
		{
			duration: time.Duration(5) * time.Second,
			rate:     10,
			ctxFunc:  context.TODO,
			output:   "\tTotal = %d\n\tSucceeded = %d\n\tFailed = %d\n\t50th percentile = %v\n\t95th percentile = %v\n\t99th percentile = %v\n\t99,9th percentile = %v\n",
		},
		// scenario 1: set context to timeout before duration
		//             clean shutdown is expected.
		{
			duration:    time.Duration(15) * time.Second,
			rate:        7,
			output:      "\tTotal = %d\n\tSucceeded = %d\n\tFailed = %d\n\t50th percentile = %v\n\t95th percentile = %v\n\t99th percentile = %v\n\t99,9th percentile = %v\n",
			ctxDuration: time.Duration(8) * time.Second,
			ctxFunc: func() context.Context {
				c, _ := context.WithTimeout(context.TODO(), time.Duration(8)*time.Second)
				return c
			},
		},
		// scenario 2: slow server responses
		{
			duration: time.Duration(15) * time.Second,
			rate:     10,
			output:   "\tTotal = %d\n\tSucceeded = %d\n\tFailed = %d\n\t50th percentile = %v\n\t95th percentile = %v\n\t99th percentile = %v\n\t99,9th percentile = %v\n",
			ctxFunc:  context.TODO,
			handler: func() {
				time.Sleep(5 * time.Second)
			},
		},
	}

	// run test scenarios
	for i, ts := range scenarios {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ts.handler != nil {
				ts.handler()
			}
			fmt.Fprintln(w, "Hello from Mars")
		}))
		defer testServer.Close()

		buf := &bytes.Buffer{}
		target, err := probe.New(ts.duration, ts.rate, testServer.URL)
		if err != nil {
			t.Errorf("scenario %d: failed to create probe object, due to %s", i, err.Error())
		}

		// start probe
		start := time.Now()
		ctx := ts.ctxFunc()
		res := target.Start(ctx)
		elapsed := time.Since(start)

		// validate samples
		if res.Total <= 0 {
			t.Errorf("scenario %d: total number of request must be > 0", i)
		}
		if res.Failed > 0 {
			t.Errorf("scenario %d: didn't expect any request to fail, failed = %d", i, res.Failed)
		}
		if res.Total != res.Succeeded+res.Failed {
			t.Errorf("sceanrio %d, succeeded + failed don't add up to total, i")
		}

		err = res.Print(buf)
		ts.output = fmt.Sprintf(
			ts.output,
			res.Total,
			res.Succeeded,
			res.Failed,
			res.P50,
			res.P95,
			res.P99,
			res.P999)
		if err != nil {
			t.Errorf("scenario %d: failed to print results, due to %s", i, err.Error())
		}
		if !strings.Contains(buf.String(), ts.output) {
			t.Errorf("scenario %d: incorrect output returned\n Got: %s\n ShouldContain: %s\n", i, buf.String(), ts.output)
		}

		// validate execution time with 1s margin
		var durationDiff time.Duration
		var expectedDuration time.Duration
		expectedDuration = ts.duration
		if ts.ctxDuration.Seconds() > 0 {
			expectedDuration = ts.ctxDuration
		}

		if elapsed > expectedDuration {
			durationDiff = elapsed - expectedDuration
		} else {
			durationDiff = expectedDuration - elapsed
		}
		if durationDiff > time.Duration(1)*time.Second {
			t.Errorf("scenario %d: probe duration exeeded\n Elapsed %v, Wanted %v, Diff %v", i, elapsed, expectedDuration, durationDiff)
		}
	}
}
