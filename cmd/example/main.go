/*
Example is a HTTP server that respondes to requests on /limited according to
the provided token bucket policy.

    $ go run example.go -limit=60 &
    $ ab -n5000 -c5 http://localhost:6060/limited

    This is ApacheBench, Version 2.3 <$Revision: 1528965 $>
    Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
    Licensed to The Apache Software Foundation, http://www.apache.org/
    
    Benchmarking localhost (be patient)
    Completed 500 requests
    Completed 1000 requests
    Completed 1500 requests
    Completed 2000 requests
    Completed 2500 requests
    Completed 3000 requests
    Completed 3500 requests
    Completed 4000 requests
    Completed 4500 requests
    Completed 5000 requests
    Finished 5000 requests
    
    
    Server Software:        
    Server Hostname:        localhost
    Server Port:            6060
    
    Document Path:          /limited
    Document Length:        0 bytes
    
    Concurrency Level:      5
    Time taken for tests:   82.328 seconds
    Complete requests:      5000
    Failed requests:        0
    Total transferred:      580000 bytes
    HTML transferred:       0 bytes
    Requests per second:    60.73 [#/sec] (mean)
    Time per request:       82.328 [ms] (mean)
    Time per request:       16.466 [ms] (mean, across all concurrent requests)
    Transfer rate:          6.88 [Kbytes/sec] received
    
    Connection Times (ms)
                  min  mean[+/-sd] median   max
    Connect:        0    0   0.0      0       0
    Processing:     0   82  73.4     66     900
    Waiting:        0   82  73.4     66     900
    Total:          0   82  73.4     66     900
    
    Percentage of the requests served within a certain time (ms)
      50%     66
      66%     83
      75%    116
      80%    133
      90%    183
      95%    233
      98%    300
      99%    350
     100%    900 (longest request)
 */
package main

import (
	"flag"
	"net/http"
	"time"

	. "github.com/matttproud/faregate"
)

func Must(c chan struct{}, err error) chan struct{} {
	if err != nil {
		panic(err)
	}
	return c
}

func main() {
	var (
		limit    = flag.Int("limit", 1, "the number of operations to allow per interval")
		interval = flag.Duration("interval", time.Second, "the interval over which limit applies")
		addr     = flag.String("addr", "localhost:6060", "the address to serve requests on")
	)
	flag.Parse()
	fg, err := New(RefreshInterval(*interval), TokenCount(uint64(*limit)))
	if err != nil {
		panic(err)
	}
	defer fg.Close()

	http.Handle("/limited", http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
		ready, _ := fg.Acquire(1)
		<-ready
		resp.WriteHeader(http.StatusOK)
	}))

	if err := http.ListenAndServe(*addr, nil); err != nil {
		panic(err)
	}
}
