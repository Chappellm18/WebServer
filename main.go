package main

import (
	//"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"strconv"
	"sync"
	//"strings"
)

type Post struct {
	ID int `json:"id"`
	Body string `json:"body"`
}

var (
	posts	= make(map[int]Post)
	nextID 	= 1
	postsMu sync.Mutex
)

func main() {
    // Define routes
    http.HandleFunc("/", isActiveHandler)

    fmt.Println("Starting up... on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func isActiveHandler(w http.ResponseWriter, r *http.Request) {

	// Only allow GET
	if r.Method != "GET" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create and serve an HTML file that lays out the endpoints
    	postsMu.Lock()
	defer postsMu.Unlock()

	content, err := ioutil.ReadFile("documentation.html")

	if err != nil {
		http.Error(w, "Unable to read HTML file", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
	//json.NewEncoder(w).Encode(content) 
}

