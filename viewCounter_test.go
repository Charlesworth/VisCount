package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

//"github.com/tsenart/vegeta/lib"

// func Test_something(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
// 	if shit() != "did it work?" { //try a unit test on function
// 		t.Error("shit did not work as expected.") // log error if it did not work as expected
// 	} else {
// 		t.Log("one test passed.") // log some info if you want
// 	}
// }

// func testrate(t *testing.T) {
//
// 	rate := uint64(100) // per second
// 	duration := 4 * time.Second
// 	targeter := vegeta.NewStaticTargeter(&vegeta.Target{
// 		Method: "GET",
// 		URL:    "http://localhost:9100/",
// 	})
// 	attacker := vegeta.NewAttacker()
//
// 	var results vegeta.Results
// 	for res := range attacker.Attack(targeter, rate, duration) {
// 		results = append(results, res)
// 	}
//
// 	metrics := vegeta.NewMetrics(results)
// 	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
// }

type MockURL struct {
	urlStr        string
	expectedWCode int
}

func TestGetHandler(t *testing.T) {

	router := httprouter.New()
	router.GET("/count/:page", countHandler)

	inputHTTP := [3]MockURL{
		//test case 0: add "test1" to the counter
		{"/count/test1", 200},
		//test case 1: add a second "test1" to the counter
		{"/count/test1", 200},
		//test case 2: add "test2" to the counter
		{"/count/test2", 200},
	}

	for i := range inputHTTP {
		w := httptest.NewRecorder()

		req, _ := http.NewRequest("GET", inputHTTP[i].urlStr, nil)

		router.ServeHTTP(w, req)
		fmt.Println(w.Code)
		if w.Code != inputHTTP[i].expectedWCode {
			t.Error("PutHandler test case", i, "returned", w.Code, "instead of", inputHTTP[i].expectedWCode)
		}
	}
}
