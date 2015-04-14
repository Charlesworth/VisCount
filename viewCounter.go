package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var BoltReadChannel = make(chan string)
var BoltWriteChannel = make(chan DataPoint)

var counter = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

var ips = struct {
	sync.RWMutex
	m map[string]bool
}{m: make(map[string]bool)}

type DataPoint struct { //possibly change to map[string]int instead of pageName and Viewcount
	PageName    string
	ViewCount   int
	UniqueViews int
}

type SavePoint struct { //test
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

		counter.RLock()
		fmt.Fprintln(w, "unique ips: ", len(ips.m))
		counter.RUnlock()

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
		tx.CreateBucketIfNotExists([]byte("m"))
		return nil
	})

	fmt.Println("bolt writer ready")

	//start a ticker for auto uploading
	ticker := time.NewTicker(time.Second * 20)

	for {
		select {
		case m := <-BoltWriteChannel:
			mjson, err := json.Marshal(m)
			errLog(err)
			boltClient.Update(func(tx *bolt.Tx) error {
				// Set the value "bar" for the key "foo".
				err = tx.Bucket([]byte("m")).Put([]byte("poo"), []byte(mjson)) //need the bucket id for this
				errLog(err)
				return nil
			})

		case <-ticker.C:
			log.Println("Tick")
			//auto upload code here
		}
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
