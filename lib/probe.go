package probe

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type voyager struct {
	ctx      context.Context
	duration time.Duration
	rate     int
	target   string
	// cancel used by Stop() method to abandon the work
	cancel context.CancelFunc
}

// New creates a brand new probe.
// The method will sanitize the input and if incorrect will return an error.
// Default values are provided for the following parameters if not set:
//   duration = 5 min
//   rate = 2
//
// TODO: Describe ctx, duration,rate, target
func New(c context.Context, d time.Duration, r int, t string) (voyager, error) {
	if r == 0 {
		fmt.Println("Rate not set, default value will be used 2")
		r = 2
	}
	if r < 1 {
		return voyager{}, errors.New("incorrect value of rate param, min >= 1")
	}
	if r > 10 {
		return voyager{}, errors.New("easy - I am not a performance tool, max rate is 10")
	}

	if d.Seconds() == 0 {
		fmt.Println("Duration not set, default value will be used 5 min")
		d = time.Duration(5) * time.Minute
	}

	//TODO: validate target
	return voyager{ctx: c, duration: d, rate: r, target: t, cancel: nil}, nil
}

//TODO: add description
func (v voyager) Start() *Samples {
	result := &Samples{store: []sample{}}
	rps := 1e9 / v.rate
	derivedCtx, cf := context.WithTimeout(v.ctx, v.duration)
	v.cancel = cf

	for {
		select {
		case <-derivedCtx.Done():
			result.calculate()
			return result
		default:
			// a unit of work done in one second
			// don't expect accuracy - exactly X number of request per second
			func() {
				ticker := time.NewTicker(time.Duration(rps))
				oneSecondCtx, _ := context.WithTimeout(derivedCtx, time.Duration(1)*time.Second)
				for {
					select {
					case <-oneSecondCtx.Done():
						ticker.Stop()
						return
					case <-ticker.C:
						result.add(sample{})
					}
				}
			}()
		}
	}
}

// Stop abandons started work.
// Does not wait for the work to complete.
func (v voyager) Stop() error {
	if v.cancel != nil {
		v.cancel()
		return nil
	}
	return errors.New("Nothing to stop, call Start() before calling Stop()")
}
