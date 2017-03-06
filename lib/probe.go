package probe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"runtime/debug"
	"sync"
	"time"
)

var (
	// urlRegex taken from https://www.owasp.org/index.php/OWASP_Validation_Regex_Repository
	urlRegex = regexp.MustCompile("(((https|http)://)(%[0-9A-Fa-f]{2}|[-()_.!~*';/?:@&=+$,A-Za-z0-9])+)([).!';/?:,][[:blank:]])?$")
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
//  duration - the period of time over which response times are gathered
//  rate - roughly the number of request per second.
//  target - a URL address of an endpoint
//
// Default values are provided for the following parameters if not set:
//   duration = 5 min
//   rate = 10
func New(duration time.Duration, rate int, target string) (voyager, error) {
	if rate == 0 {
		rate = 10
		fmt.Printf("rate not set, default value will be used %d\n", rate)
	}
	if rate < 1 {
		return voyager{}, errors.New("incorrect value of rate param, min >= 1")
	}
	if rate > 10 {
		return voyager{}, errors.New("easy - I am not a performance tool, max rate is 10")
	}

	if duration.Seconds() == 0 {
		duration = time.Duration(5) * time.Minute
		fmt.Printf("duration not set, default value will be used %v", duration)
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

	if !urlRegex.MatchString(target) {
		err := fmt.Errorf("incorrect url passed in, target = %s", target)
		return voyager{}, err
	}
	if _, err := url.ParseRequestURI(target); err != nil {
		err := fmt.Errorf("incorrect url passed in, target = %s, err = %s", target, err.Error())
		return voyager{}, err
	}
	return voyager{duration: duration, rate: rate, target: target, hClient: client}, nil
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

func (v voyager) workerWrapper(ctx context.Context, resCh chan<- sample, wg *sync.WaitGroup) {
	defer wg.Done()
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
