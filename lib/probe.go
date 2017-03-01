package probe

import (
	"context"
	"time"
)

type voyager struct{}

//TODO: add description
func New() voyager {
	return voyager{}
}

//TODO: add description
func (v voyager) Start(_ context.Context, _ time.Duration, _ int, target string) Printer {
	data := samples{store: []sample{}}
	return data
}
