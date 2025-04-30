package main

import (
	//"encoding/json"
	"fmt"
	//"os"
	"io/ioutil"
	"log"
	"net/http"
	//"strconv"
	"sync"
	//"runtime"
	"strings"
)

type Post struct {
	ID int `json:"id"`
	Body string `json:"body"`
}

type Board struct {
    Name     string `json:"name,omitempty"`
    Vendor   string `json:"vendor,omitempty"`
    Version  string `json:"version,omitempty"`
    Serial   string `json:"serial,omitempty"`
    AssetTag string `json:"assettag,omitempty"`
}

var posts []string
var postsMu sync.Mutex

var docContent []byte
var docContentMu sync.Mutex

func main() {
	// Define routes
    	http.HandleFunc("/", activeHandler)
	http.HandleFunc("/cpu", cpuUsageHandler)
	http.HandleFunc("/mem", memUsageHandler)
	http.HandleFunc("/hardware", hardwareHandler)

    	fmt.Println("Starting up... on port 8080")
    	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Read HTML to serve
func readHTML(filename string) []byte {
	docContentMu.Lock()
	defer docContentMu.Unlock()

	tempContent, err := ioutil.ReadFile("templates/" + filename)
	if err != nil { log.Fatal(err) }
	return tempContent
}
// Default Route
func activeHandler(w http.ResponseWriter, r *http.Request) {	
	docContent = readHTML("documention.html")
	
	w.Header().Set("Content-Type", "text/html")
	w.Write(docContent) 
}

// CPU Usage Route
func cpuUsageHandler(w http.ResponseWriter, r *http.Request) {
	docContent = readHTML("cpu.html")
	
	w.Header().Set("Content-Type", "text/html")
	w.Write(docContent)
}

// MEM Usage Route
func memUsageHandler(w http.ResponseWriter, r *http.Request) {
	docContent = readHTML("mem.html")

	w.Header().Set("Content-Type", "text/html")
	w.Write(docContent)
}

// Hardware Route
func hardwareHandler(w http.ResponseWriter, r *http.Request) {
	hardwareInfo, err := boardInfo()
	if err != nil {
        	fmt.Println("Error getting board info:", err)
        	return
    	}
	
    	fmt.Printf("Full Board struct: %+v\n", hardwareInfo)

	docContent = readHTML("cpu.html")

	w.Header().Set("Content-Type", "text/html")
	w.Write(docContent)
}

func boardInfo() (Board, error) {
    var b Board
    var err error

    if b.Name, err = slurpFile("/sys/class/dmi/id/board_name"); err != nil {
        return b, fmt.Errorf("reading board name: %w", err)
    }
    if b.Vendor, err = slurpFile("/sys/class/dmi/id/board_vendor"); err != nil {
        return b, fmt.Errorf("reading board vendor: %w", err)
    }
    if b.Version, err = slurpFile("/sys/class/dmi/id/board_version"); err != nil {
        return b, fmt.Errorf("reading board version: %w", err)
    }
    if b.Serial, err = slurpFile("/sys/class/dmi/id/board_serial"); err != nil {
        return b, fmt.Errorf("reading board serial: %w", err)
    }
    if b.AssetTag, err = slurpFile("/sys/class/dmi/id/board_asset_tag"); err != nil {
        return b, fmt.Errorf("reading board asset tag: %w", err)
    }

    return b, nil
}
func slurpFile(path string) (string, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(data)), nil
}













