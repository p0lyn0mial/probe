package probe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type voyager struct {
	duration time.Duration
	rate     int
	target   string
	hClient  *http.Client
}

// New creates a brand new probe.
// The method will sanitize the input and if incorrect will return an error.
//
// Arguments:
//  d - duration, the period of time over which response times are gathered
//  r - rate, roughly the number of request per second.
//  t - target, a URL address of an endpoint
//
// Default values are provided for the following parameters if not set:
//   duration = 5 min
//   rate = 10
func New(d time.Duration, r int, t string) (voyager, error) {
	if r == 0 {
		r = 10
		fmt.Printf("Rate not set, default value will be used %d\n", r)
	}
	if r < 1 {
		return voyager{}, errors.New("incorrect value of rate param, min >= 1")
	}
	if r > 10 {
		return voyager{}, errors.New("easy - I am not a performance tool, max rate is 10")
	}

	if d.Seconds() == 0 {
		d = time.Duration(5) * time.Minute
		fmt.Printf("Duration not set, default value will be used %v", d)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			MaxIdleConns:        20,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: 60 * time.Second,
	}
	//TODO: validate target
	return voyager{duration: d, rate: r, target: t, hClient: client}, nil
}

//TODO: add description
func (v voyager) Start(ctx context.Context) *Samples {
	var wg sync.WaitGroup
	results := &Samples{rspTimes: []float64{}}
	resultsCh := make(chan sample)
	// synchronization primitive between consumer and main loop
	consumerDoneCh := make(chan bool)
	defer close(resultsCh)
	defer close(consumerDoneCh)

	rps := 1e9 / v.rate
	derivedCtx, _ := context.WithTimeout(ctx, v.duration)

	// consumer
	// if this guy goes down there is a deadlock
	// TODO: maybe is is better to let it panic
	go func() {
		defer handleCrash("Unexpected error occured, please kill the process and restart the app")
		for {
			select {
			case s := <-resultsCh:
				results.add(s)
			case <-consumerDoneCh:
				return
			}
		}
	}()

	// main loop
	for {
		select {
		case <-derivedCtx.Done():
			wg.Wait()
			consumerDoneCh <- true
			results.calculate()
			return results
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
						wg.Add(1)
						go v.workerWrapper(derivedCtx, resultsCh, &wg)
					}
				}
			}()
		}
	}
}

func (v voyager) workerWrapper(ctx context.Context, resCh chan<- sample, changeName *sync.WaitGroup) {
	defer changeName.Done()
	defer handleCrash("TODO: give me a message")
	rspTime, statusCode, err := v.worker(ctx)
	if err == nil {
		s := sample{
			rspTime:    rspTime,
			statusCode: statusCode,
		}
		resCh <- s
	}
}

// worker makes actual HTTP GET request to target
func (v voyager) worker(ctx context.Context) (time.Duration, int, error) {
	req, err := http.NewRequest("GET", v.target, nil)
	if err != nil {
		return 0, 0, err
	}
	req = req.WithContext(ctx)

	start := time.Now()
	res, err := v.hClient.Do(req)
	//TODO: is this the best place to calculate elapsed time ??!!
	//maybe after reading the body ?
	elapsed := time.Since(start)
	if err != nil {
		return 0, 0, err
	}

	defer res.Body.Close()
	_, err = io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		return 0, 0, err
	}
	return elapsed, res.StatusCode, nil
}

// handleCrash simply catches a crash and prints an error. Meant to be called via defer
func handleCrash(msg string) {
	if r := recover(); r != nil {
		err := fmt.Errorf("caught the following panic message: %v, stack: %s", r, debug.Stack())
		fmt.Println(err)
	}
}
