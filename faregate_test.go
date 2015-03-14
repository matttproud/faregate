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

// bmFaregate provides a crude throughput measure for the faregate.
func bmFaregate(b *testing.B, m int) {
	fg, err := New(RefreshInterval(time.Millisecond), TokenCount(uint64(m*b.N)))
	if err != nil {
		b.Fatalf("New() = _, %s; want _, <nil>", err)
	}
	defer fg.Close()
	var start, end sync.WaitGroup
	start.Add(1)
	for i := 0; i < m; i++ {
		end.Add(1)
		go func() {
			start.Wait()
			for j := 0; j < b.N; j++ {
				ready, _ := fg.Acquire(1)
				<-ready
			}
			end.Done()
		}()
	}
	b.ResetTimer()
	b.ReportAllocs()
	start.Done()
	end.Wait()
}

func BenchmarkFaregate1(b *testing.B) {
	bmFaregate(b, 1)
}

func BenchmarkFaregate2(b *testing.B) {
	bmFaregate(b, 2)
}

func BenchmarkFaregate4(b *testing.B) {
	bmFaregate(b, 4)
}

func BenchmarkFaregate8(b *testing.B) {
	bmFaregate(b, 8)
}
