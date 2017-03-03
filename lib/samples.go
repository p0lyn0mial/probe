package probe

import (
	"fmt"
	"io"
	"time"
)

//TODO: add description
type sample struct {
	rspTime time.Duration
	succeed bool
}

//TODO: add description
type Samples struct {
	wasCalculated bool
	store         []sample
	Total         int
	Failed        int
	Succeeded     int
}

//TODO: add description
func (ss *Samples) add(s sample) {
	ss.store = append(ss.store, s)
}

//TODO: add description
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
	return nil
}

func (ss *Samples) calculate() {
	if ss.wasCalculated {
		return
	}
	for _, s := range ss.store {
		ss.Total++
		if s.succeed {
			ss.Succeeded++
		} else {
			ss.Failed++
		}
	}
	ss.wasCalculated = true
}
