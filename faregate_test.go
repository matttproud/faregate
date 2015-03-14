// Package faregate provides a token bucket load shaper.
package faregate

import (
	"sync"
	"testing"
	"time"
)

func TestFaregate(t *testing.T) {
	fg, err := New(RefreshInterval(time.Millisecond), TokenCount(10))
	if err != nil {
		t.Fatalf("New() = _, %s; want _, <nil>", err)
	}
	defer fg.Close()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ready, err := fg.Acquire(1)
			if err != nil {
				t.Fatalf("fg.Acquire(1) = _, %s; want _, <nil>", err)
			}
			select {
			case <-ready:
				return
			case <-time.After(time.Millisecond):
				t.Fatalf("ready timed out")
			}
		}()
	}
	wg.Wait()
}

func TestFaregateBacklog(t *testing.T) {
	fg, err := New(RefreshInterval(time.Millisecond), TokenCount(10))
	if err != nil {
		t.Fatalf("New() = _, %s; want _, <nil>", err)
	}
	defer fg.Close()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ready, err := fg.Acquire(2)
			if err != nil {
				t.Fatalf("fg.Acquire(1) = _, %s; want _, <nil>", err)
			}
			select {
			case <-ready:
				return
			case <-time.After(3 * time.Millisecond):
				t.Fatalf("ready timed out")
			}
		}()
	}
	wg.Wait()
}
