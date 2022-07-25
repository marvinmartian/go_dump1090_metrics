// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AircraftList struct {
	Messages int32      `json:"messages"`
	Aircraft []Aircraft `json:"aircraft"`
}

type Statistics struct {
	Latest  SingleStat `json:"latest"`
	Last_1  SingleStat `json:"last1min"`
	Last_5  SingleStat `json:"last5min"`
	Last_15 SingleStat `json:"last15min"`
	Total   SingleStat `json:"total"`
}

type SingleStat struct {
	Start float32 `json:"start"`
	End   float32 `json:"end"`
}

type Aircraft struct {
	Hex         string  `json:"hex"`
	Flight      string  `json:"flight,omitempty"`
	AltoBaro    uint16  `json:"alt_baro,omitempty"`
	GroundSpeed float32 `json:"gs,omitempty"`
	BaroRate    int16   `json:"baro_rate,omitempty"`
	Latitude    float32 `json:"lat,omitempty"`
	Longitude   float32 `json:"lon,omitempty"`
	NavHeading  float32 `json:"nav_heading,omitempty"`
	RSSi        float32 `json:"rssi,omitempty"`
	Category    string  `json:"category,omitempty"`
	Emergency   string  `json:"emergency,omitempty"`
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func readAircraftFile(path string) {

	// Open the file
	jsonFile, err := os.Open(path + "aircraft.json")

	// Print the error if that happens.
	if err != nil {
		fmt.Println(err)
	}

	// defer file close
	defer jsonFile.Close()

	// read file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// Initialize list of aircraft
	var aircraft_list AircraftList

	// Unmarshal to aircraft list
	json.Unmarshal(byteValue, &aircraft_list)

	fmt.Printf("%+v\n", aircraft_list)
}

func readStatsFile(path string) {

	// Open the file
	jsonFile, err := os.Open(path + "stats.json")

	// Print the error if that happens.
	if err != nil {
		fmt.Println(err)
	}

	// defer file close
	defer jsonFile.Close()

	// read file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// Initialize list of aircraft
	var stats Statistics

	// Unmarshal to aircraft list
	json.Unmarshal(byteValue, &stats)

	fmt.Printf("%+v\n", stats)
}

func readFiles(path string) {
	ticker := time.NewTicker(5 * time.Second)

	for _ = range ticker.C {
		// fmt.Println("Ticking..")
		opsProcessed.Inc()
		readAircraftFile(path)
		readStatsFile(path)
	}

}

func main() {

	path := flag.String("path", "/run/dump1090-fa/", "Path to json files. Default /run/dump1090-fa/")
	flag.Parse()
	fmt.Println(*path)

	go readFiles(*path)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)

}
