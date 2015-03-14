package faregate

import (
	"fmt"
	"math/rand"
	"time"
)

func Must(c <-chan struct{}, err error) <-chan struct{} {
	if err != nil {
		panic(err)
	}
	return c
}

func Example() {
	rnd := rand.New(rand.NewSource(42))

	fg, err := New(RefreshInterval(time.Second), TokenCount(100), ConcurrencyLevel(1))
	if err != nil {
		panic(err)
	}
	defer fg.Close()

	for {
		q := rnd.Intn(10)
		<-Must(fg.Acquire(uint64(q)))
		fmt.Println("acquired", q)
	}
}
