package probe

import (
	"fmt"
	"io"
)

//TODO: add description
type sample struct{}

//TODO: add description
type samples struct {
	store []sample
}

//TODO: add description
type Printer interface {
	Print(io.Writer) error
}

//TODO: add description
func (ss samples) add(s sample) {
	ss.store = append(ss.store, s)
}

//TODO: add description
func (ss samples) Print(w io.Writer) error {
	_, err := fmt.Fprint(w, "Printing collected samples")
	return err
}

var _ Printer = samples{nil}
