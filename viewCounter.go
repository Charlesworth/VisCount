package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var BoltReadChannel = make(chan string)

//var BoltWriteChannel = make(chan DataPoint)

var counter = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

var ips = struct {
	sync.RWMutex
	m map[string]bool
}{m: make(map[string]bool)}

type IPList struct {
	IPs map[string]bool
}

type SavePoint struct {
	PageCounts  map[string]int
	UniqueViews int
}

func main() {

	procNo := runtime.NumCPU()
	runtime.GOMAXPROCS(procNo)
	fmt.Println("Using", procNo, "processors for maximum thread count")

	go boltWriteClient()

	router := httprouter.New()
	router.GET("/count/:pageID", countHandler)
	router.GET("/stats/:pswrd", statsHandler)

	http.Handle("/", router)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func countHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println(r.RemoteAddr + " requests " + params.ByName("pageID"))

	counter.Lock()
	counter.m[params.ByName("pageID")]++
	counter.Unlock()

	ips.Lock()
	ips.m[r.RemoteAddr] = true
	ips.Unlock()
}

func statsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("stat request from " + r.RemoteAddr)
	if params.ByName("pswrd") == "opensaysame" {

		fmt.Fprintln(w, "Hi there, heres your stats")

		counter.RLock()
		fmt.Fprintln(w, counter.m)
		counter.RUnlock()

		ips.RLock()
		fmt.Fprintln(w, "unique ips: ", len(ips.m))
		ips.RUnlock()
	} else {
		fmt.Fprintln(w, "access denied")
	}
}

func errFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func errLog(err error) {
	if err != nil {
		log.Print(err)
	}
}

func boltWriteClient() {
	boltClient, err := bolt.Open("viewCounter.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer boltClient.Close()

	//do we need to check a bucket exists or make one
	boltClient.Update(func(tx *bolt.Tx) error {
		// Create a bucket.
		tx.CreateBucketIfNotExists([]byte("historicData"))
		return nil
	})

	fmt.Println("bolt writer ready")

	//start a ticker for auto uploading
	ticker := time.NewTicker(time.Hour)

	for {

		<-ticker.C
		log.Println("Tick")

		date := strconv.Itoa((time.Now().YearDay() * 10000) + time.Now().Year())
		fmt.Println(date)

		counter.RLock()
		ips.RLock()

		m1 := SavePoint{
			PageCounts:  counter.m,
			UniqueViews: len(ips.m),
		}

		m2 := IPList{
			IPs: ips.m,
		}

		counter.RUnlock()
		ips.RUnlock()

		m1json, err := json.Marshal(m1)
		errLog(err)
		m2json, err := json.Marshal(m2)
		errLog(err)
		boltClient.Update(func(tx *bolt.Tx) error {

			err = tx.Bucket([]byte("historicData")).Put([]byte(date), []byte(m1json))
			errLog(err)

			err = tx.Bucket([]byte("historicData")).Put([]byte("IPs"), []byte(m2json))
			errLog(err)
			return nil
		})

	}
}

// func boltReadClient() {
// 	boltClient, err := bolt.Open("viewCounter.db", 0600, nil) //maybe change the 600 to a read only value
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer boltClient.Close()

// 	fmt.Println("bolt reader ready")

// 	for {
// 		id := <-BoltReadChannel

// 		var b []byte
// 		boltClient.View(func(tx *bolt.Tx) error {
// 			// Set the value "bar" for the key "foo".
// 			b = tx.Bucket([]byte("m")).Get([]byte(id))
// 			errLog(err)

// 			return nil
// 		})

// 		//var mjson Message
// 		//err := json.Unmarshal(b, &mjson)

// 	}
// }
