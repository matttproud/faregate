// Package faregate provides a token bucket load shaper.
package faregate

import (
	"errors"
	"time"
)

// option captures the accumulated state for the Faregate's configuration.
type option struct {
	TokenCount      uint64
	RefreshInterval time.Duration
}

//optionFn applies a configuration to the option.
type optionFn func(*option) error

// RefreshInterval sets how frequently the bucket's tokens are refreshed.
func RefreshInterval(i time.Duration) optionFn {
	return func(o *option) error {
		o.RefreshInterval = i
		return nil
	}
}

var errIllegalTokenCnt = errors.New("faregate: illegal token count")

// TokenCount sets how large the bucket is.
func TokenCount(c uint64) optionFn {
	return func(o *option) error {
		if c <= 0 {
			return errIllegalTokenCnt
		}
		o.TokenCount = c
		return nil
	}
}

var errIllegalRefreshInt = errors.New("faregate: illegal refresh interval")

func newOption(opts ...optionFn) (*option, error) {
	var opt option
	for _, o := range opts {
		if err := o(&opt); err != nil {
			return nil, err
		}
	}
	if opt.TokenCount == 0 {
		return nil, errIllegalTokenCnt
	}
	if opt.RefreshInterval == 0 {
		return nil, errIllegalRefreshInt
	}
	return &opt, nil
}

// Faregate is a token bucket load shaper.
type Faregate struct {
	refreshQuantity uint64        // no. tokens when refreshed.
	quantity        uint64        // no. tokens left before refresh.
	tick            *time.Ticker  // updater for token refresh.
	ops             chan *op      // token request operations.
	quit            chan struct{} // termination channel.
}

// New creates a Faregate from the options.
func New(opts ...optionFn) (*Faregate, error) {
	opt, err := newOption(opts...)
	if err != nil {
		return nil, err
	}
	gate := &Faregate{
		refreshQuantity: opt.TokenCount,
		quantity:        opt.TokenCount,
		tick:            time.NewTicker(opt.RefreshInterval),
		ops:             make(chan *op),
		quit:            make(chan struct{}),
	}
	go gate.serve()
	return gate, nil
}

// op captures a request for tokens.
type op struct {
	count uint64        // how many tokens.
	ready chan struct{} // indicator for when tokens have been acquired..
}

// run is the Faregate's event loop.
func (f *Faregate) serve() {
	defer f.tick.Stop()
	backlog := make(chan *op, 1)
	go func() {
		var (
			queue    []*op    // accumulation of pending ops if insufficient tokens.
			backChan chan *op // points to f.ops only if len(queue) > 0.
		)
		for {
			switch {
			case backChan == nil:
				queue = append(queue, <-backlog)
				backChan = f.ops
			default:
				select {
				case b := <-backlog:
					queue = append(queue, b)
				case backChan <- queue[0]:
					queue = queue[1:]
					if len(queue) == 0 {
						backChan = nil
					}
				}
			}
		}
	}()
	for {
		select {
		case <-f.quit:
			return
		case <-f.tick.C:
			f.quantity = f.refreshQuantity
		case r := <-f.ops:
			switch {
			case r.count > f.quantity:
				backlog <- r
			default:
				f.quantity -= r.count
				close(r.ready)
			}
		}
	}
}

// Close stops the Faregate.
func (f *Faregate) Close() {
	close(f.quit)
}

var errTooManyTickets = errors.New("faregate: too many tickets requested")

// Acquire acquires n tokens from the faregate.  It returns a channel whose
// closure indicates when tokens were acquired.  Token requests are typically
// answered fairly.  If there is backlogging of tokens requests, the fairness
// guarantees break down.
func (f *Faregate) Acquire(n uint64) (chan struct{}, error) {
	if n > f.refreshQuantity {
		return nil, errTooManyTickets
	}
	ready := make(chan struct{})
	f.ops <- &op{n, ready}
	return ready, nil
}
