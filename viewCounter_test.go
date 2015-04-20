package main

import (
	"fmt"
	"github.com/tsenart/vegeta"
	"testing"
	"time"
)

func Test_something(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	if shit() != "did it work?" { //try a unit test on function
		t.Error("shit did not work as expected.") // log error if it did not work as expected
	} else {
		t.Log("one test passed.") // log some info if you want
	}
}

func testrate(t *testing.T) {

	rate := uint64(100) // per second
	duration := 4 * time.Second
	targeter := vegeta.NewStaticTargeter(&vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:9100/",
	})
	attacker := vegeta.NewAttacker()

	var results vegeta.Results
	for res := range attacker.Attack(targeter, rate, duration) {
		results = append(results, res)
	}

	metrics := vegeta.NewMetrics(results)
	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}
