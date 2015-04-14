package main

import (
	//"encoding/json"
	"fmt"
	//"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"runtime"
	"sync"
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

type DataPoint struct {
	PageName    string
	ViewCount   int
	UniqueViews int
}

//buckets = days, new bucket for each day
//key = page name, value = cumulative views
//key = unique views, value = unique ids that day

//maybe use an atomic counter for page veiws or a mutex

//maybe use a map implemented as a set for the list of IPs
//where the key is the ip and the value is anything (bool for small size)
//you can use map len to find unique views for the day

//bucket = unique ids
//set a count at the start of a day, the increase by the end of the day is the unique views value from above

//use a ticker to signal storing of the days bucket

func main() {

	procNo := runtime.NumCPU()
	runtime.GOMAXPROCS(procNo)
	fmt.Println("Using", procNo, "processors for maximum thread count")

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

	// counter.RLock()
	// n := counter.m[params.ByName("pageID")]
	// counter.RUnlock()
	// log.Println(params.ByName("pageID"), n)

	ips.Lock()
	ips.m[r.RemoteAddr] = true
	ips.Unlock()
}

func statsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("stat request from " + r.RemoteAddr)
	if params.ByName("pswrd") == "opensaysame" {
		fmt.Fprintln(w, "Hi there, heres your stats")
		//read DB
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

// func boltWriteClient() {
// 	boltClient, err := bolt.Open("viewCounter.db", 0600, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer boltClient.Close()

// 	//do we need to check a bucket exists or make one
// 	boltClient.Update(func(tx *bolt.Tx) error {
// 		// Create a bucket.
// 		tx.CreateBucketIfNotExists([]byte("m"))
// 		return nil
// 	})

// 	fmt.Println("bolt writer ready")

// 	for {

// 		m := <-BoltWriteChannel
// 		mjson, err := json.Marshal(m)
// 		errLog(err)
// 		boltClient.Update(func(tx *bolt.Tx) error {
// 			// Set the value "bar" for the key "foo".
// 			err = tx.Bucket([]byte("m")).Put([]byte("poo"), []byte(mjson)) //need the bucket id for this
// 			errLog(err)
// 			return nil
// 		})

// 	}
// }

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
