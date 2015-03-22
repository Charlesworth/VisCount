package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
)

var C chan string = make(chan string)

func main() {

	db := startDB()
	defer db.Close()

	go writeToDB(C, db)

	router := httprouter.New()
	router.GET("/count/:pageID", countHandler)
	router.GET("/stats/:pswrd", statsHandler)

	http.Handle("/", router)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func countHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println(r.RemoteAddr + " requests " + params.ByName("pageID"))
	C <- params.ByName("pageID")
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

//BoltDB can only have one writ connection at a time, so a single goRoutine#
//manages all of the writes. The page ID is send via channel C and it then
//checks if already there in the DB and writes a new view count value depending.
func writeToDB(c chan string, db *bolt.DB) {
	//Infinite Loop
	for {
		select {
		//when we recieve a message from the channel C
		case msg1 := <-c:
			fmt.Println("message received: " + msg1)
			//open a read/write connection with the DB
			db.Update(func(tx *bolt.Tx) error {
				//find Key msg1 and retrieve the value
				value := tx.Bucket([]byte("viewCount")).Get([]byte(msg1))
				s, _ := strconv.Atoi(string(value))
				fmt.Println("incoming value ", s)

				//If key returned nil then write the first value = 1
				if value == nil {
					tx.Bucket([]byte("viewCount")).Put([]byte(msg1), []byte("1"))
					fmt.Println("first time")
					return nil

					//else read the value, and write value++
				} else {
					g := s + 1
					tx.Bucket([]byte("viewCount")).Put([]byte(msg1), []byte(string(g)))
					fmt.Println(g)
					return nil
				}
			})
		}
	}
}

func startDB() *bolt.DB {
	db, err := bolt.Open("veiwCount.db", 0600, nil)
	errFatal(err)

	db.Update(func(tx *bolt.Tx) error {
		// Create a bucket.
		tx.CreateBucketIfNotExists([]byte("viewCount"))

		return nil
	})

	return db
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
