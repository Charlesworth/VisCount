package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

//Instance of a [pageName]pageView hash map. This is implemented with a
//mutex RW lock to stop goroutine data races
var counter = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

//Instance of a [ipAdress]bool hash map. We don't care about the bool,
//using a hash map in this case just for the IP Key, as it offers a
//easy implementation on a set with quick insertion. This struct has a
//mutex RW lock to stop goroutine data races
var ips = struct {
	sync.RWMutex
	m map[string]bool
}{m: make(map[string]bool)}

//IPList struct is used to marshal/unmarshal IP visitor data into JSON
//to be sent to current storage
type IPList struct {
	IPs map[string]bool
}

//SavePoint struct is used to marshal/unmarshal pageview data into JSON
//to be sent to current and historic storage
type SavePoint struct {
	PageCounts  map[string]int
	UniqueViews int
}

//Main checks checks for previos data, sets up multithreading and then
//initiates the HTTP server
func main() {

	//checks for present DB storage and loads it into memory
	checkForRecords()

	//find the amount of available cores and set the runtime to
	//utalize all of them
	procNo := runtime.NumCPU()
	runtime.GOMAXPROCS(1)
	fmt.Println("Using", procNo, "processors for maximum thread count")

	go periodicMemoryWriter()

	router := httprouter.New()
	router.GET("/count/:pageID", countHandler)
	router.GET("/stats/:pswrd", statsHandler)

	http.Handle("/", router)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

//
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

func periodicMemoryWriter() {
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
	ticker := time.NewTicker(time.Second * 10) //time.Hour)

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

			err = tx.Bucket([]byte("historicData")).Put([]byte("current"), []byte(m1json))
			errLog(err)

			err = tx.Bucket([]byte("historicData")).Put([]byte("IPs"), []byte(m2json))
			errLog(err)
			return nil
		})

	}
}

func checkForRecords() {
	if _, err := os.Stat("viewCounter.db"); err == nil {
		log.Println("viewCount.db database already exists; processing old entries")

		boltClient, err := bolt.Open("viewCounter.db", 0600, nil) //maybe change the 600 to a read only value
		if err != nil {
			log.Fatal(err)
		}
		defer boltClient.Close()

		// fmt.Println("bolt reader ready")

		// //id := <-BoltReadChannel
		log.Println("point 0")
		var b1, b2 []byte
		boltClient.View(func(tx *bolt.Tx) error {
			// Set the value "bar" for the key "foo".
			b1 = tx.Bucket([]byte("historicData")).Get([]byte("current"))
			errLog(err)

			b2 = tx.Bucket([]byte("historicData")).Get([]byte("IPs"))
			errLog(err)

			return nil
		})

		var mjson1 SavePoint
		err = json.Unmarshal(b1, &mjson1)
		errLog(err)

		for k, v := range mjson1.PageCounts {
			counter.m[k] = v
		}

		log.Println("point 1")
		log.Println("unique views", mjson1.UniqueViews)
		log.Println(mjson1.PageCounts["wee"])

		var mjson2 IPList
		err = json.Unmarshal(b2, &mjson2)
		errLog(err)

		log.Println("point 2")
		log.Println("unique IPs", len(mjson2.IPs))

		for k, _ := range mjson2.IPs {
			ips.m[k] = true
		}

	} else {
		log.Println("viewCount.db not present; creating database")

	}
}
