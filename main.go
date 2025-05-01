package main

import (
	//"encoding/json"
	"fmt"
	//"os"
	"io/ioutil"
	"log"
	"net/http"
	//"strconv"
	"html/template"
	"path/filepath"
	"sync"
	//"runtime"
	"strings"
	"time"
	"strconv"
)

type MemInfo struct {
    Total       uint64
    Available   uint64
    Used        uint64
    UsedPercent float64
}

type CPUInfo struct {
    Usage float64
}

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

type Route struct {
    Method      string
    Path        string
    Description string
}

type PageData struct {
    Title   string
    Heading string
    Routes  []Route
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
func renderTemplate(w http.ResponseWriter, filename string, data interface{}) {
	docContentMu.Lock()
	defer docContentMu.Unlock()

	// Safely build full path
	path := filepath.Join("templates", filename)

	// Parse the template file
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		log.Printf("error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the template with provided data
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
// Default Route
func activeHandler(w http.ResponseWriter, r *http.Request) {	
	data := PageData{
		Title:   "API Docs",
		Heading: "System Information API Docs",
		Routes: []Route{
			{"GET", "/", "Documentation endpoint"},
			{"GET", "/mem", "System memeory information"},
			{"GET", "/cpu", "System CPU information"},
			{"GET", "/hardware", "System hardware information"},
		},
	}
	renderTemplate(w, "documentation.html", data)
}

// CPU Usage Route
func cpuUsageHandler(w http.ResponseWriter, r *http.Request) {
    info, err := cpuInfo()
    if err != nil {
        fmt.Println("Error getting CPU info:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    renderTemplate(w, "cpu.html", info)
}

// MEM Usage Route
func memUsageHandler(w http.ResponseWriter, r *http.Request) {
    info, err := memInfo()
    if err != nil {
        fmt.Println("Error getting memory info:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    renderTemplate(w, "mem.html", info)
}

// Hardware Route
func hardwareHandler(w http.ResponseWriter, r *http.Request) {
    hardwareInfo, err := boardInfo()
    if err != nil {
        fmt.Println("Error getting board info:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    fmt.Printf("Full Board struct: %+v\n", hardwareInfo)

    renderTemplate(w, "hardware.html", hardwareInfo) 
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

func cpuInfo() (CPUInfo, error) {
    idle0, total0, err := readCPUTimes()
    if err != nil {
        return CPUInfo{}, err
    }

    time.Sleep(1 * time.Second)

    idle1, total1, err := readCPUTimes()
    if err != nil {
        return CPUInfo{}, err
    }

    idleDelta := idle1 - idle0
    totalDelta := total1 - total0

    usage := 100.0 * (1.0 - float64(idleDelta)/float64(totalDelta))
    return CPUInfo{Usage: usage}, nil
}

func readCPUTimes() (idle, total uint64, err error) {
    data, err := ioutil.ReadFile("/proc/stat")
    if err != nil {
        return
    }

    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "cpu ") {
            fields := strings.Fields(line)
            var values []uint64
            for _, field := range fields[1:] {
                v, err := strconv.ParseUint(field, 10, 64)
                if err != nil {
                    return 0, 0, err
                }
                values = append(values, v)
            }

            idle = values[3] // idle
            for _, val := range values {
                total += val
            }
            return
        }
    }

    return 0, 0, fmt.Errorf("cpu line not found in /proc/stat")
}



func memInfo() (MemInfo, error) {
    data, err := ioutil.ReadFile("/proc/meminfo")
    if err != nil {
        return MemInfo{}, err
    }

    lines := strings.Split(string(data), "\n")
    mem := make(map[string]uint64)

    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) < 2 {
            continue
        }
        key := strings.TrimSuffix(fields[0], ":")
        value, err := strconv.ParseUint(fields[1], 10, 64)
        if err != nil {
            continue
        }
        mem[key] = value // in KB
    }

    total := mem["MemTotal"]
    available := mem["MemAvailable"]
    used := total - available
    usedPercent := float64(used) / float64(total) * 100

    return MemInfo{
        Total:       total * 1024,    // convert to bytes
        Available:   available * 1024,
        Used:        used * 1024,
        UsedPercent: usedPercent,
    }, nil
}








