package probe

import (
	"fmt"
	"io"
)

//TODO: add description
type sample struct{}

//TODO: add description
type Samples struct {
	wasCalculated bool
	store         []sample
	Total         int
}

//TODO: add description
func (ss *Samples) add(s sample) {
	ss.store = append(ss.store, s)
}

//TODO: add description
func (ss *Samples) Print(w io.Writer) error {
	if !ss.wasCalculated {
		ss.calculate()
	}
	_, err := fmt.Fprintf(w, fmt.Sprintf("\tTotal = %d\n", ss.Total))
	return err
}

func (ss *Samples) calculate() {
	if ss.wasCalculated {
		return
	}
	ss.Total = len(ss.store)
	ss.wasCalculated = true
}
