package probe

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/montanaflynn/stats"
)

//TODO: add description
type sample struct {
	rspTime    time.Duration
	statusCode int
}

//TODO: add description
type Samples struct {
	wasCalculated bool
	rspTimes      []float64
	// Total the total number of requests
	Total int
	// Failed the number of failed requests
	Failed int
	// Succeeded the number of succeeded requests
	Succeeded int
	// P50 median - 50th percentile
	P50 time.Duration
	// P95 95th percentile
	P95 time.Duration
	// P99 99th percentile
	P99 time.Duration
	// P999 99,9th percentile
	P999 time.Duration
}

//TODO: add description
func (ss *Samples) add(s sample) {
	if s.statusCode == http.StatusOK || s.statusCode == http.StatusNoContent {
		ss.Succeeded++
	} else {
		ss.Failed++
	}
	ss.rspTimes = append(ss.rspTimes, float64(s.rspTime))
}

//TODO: add description
//TODO: for pretty print use text/tabwriter from std lib
func (ss *Samples) Print(w io.Writer) (err error) {
	if !ss.wasCalculated {
		ss.calculate()
	}
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	_, err = fmt.Fprintf(w, fmt.Sprintf("\tTotal = %d\n", ss.Total))
	checkErr(err)
	_, err = fmt.Fprintf(w, fmt.Sprintf("\tSucceeded = %d\n", ss.Succeeded))
	checkErr(err)
	_, err = fmt.Fprintf(w, fmt.Sprintf("\tFailed = %d\n", ss.Failed))
	checkErr(err)
	_, err = fmt.Fprintf(w, fmt.Sprintf("\t50th percentile = %v\n", ss.P50))
	checkErr(err)
	_, err = fmt.Fprintf(w, fmt.Sprintf("\t95th percentile = %v\n", ss.P95))
	checkErr(err)
	_, err = fmt.Fprintf(w, fmt.Sprintf("\t99th percentile = %v\n", ss.P99))
	checkErr(err)
	_, err = fmt.Fprintf(w, fmt.Sprintf("\t99,9th percentile = %v\n", ss.P999))
	checkErr(err)
	return nil
}

func (ss *Samples) calculate() {
	if ss.wasCalculated {
		return
	}
	ss.Total = ss.Succeeded + ss.Failed
	if p50, err := stats.Percentile(ss.rspTimes, 50); err == nil {
		ss.P50 = time.Duration(p50)
	}
	if p95, err := stats.Percentile(ss.rspTimes, 95); err == nil {
		ss.P95 = time.Duration(p95)
	}
	if p99, err := stats.Percentile(ss.rspTimes, 99); err == nil {
		ss.P99 = time.Duration(p99)
	}
	if p999, err := stats.Percentile(ss.rspTimes, 99.9); err == nil {
		ss.P999 = time.Duration(p999)
	}
	ss.wasCalculated = true
}
