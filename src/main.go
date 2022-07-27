// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
	AltoGeom    uint16  `json:"alt_geom,omitempty"`
	GroundSpeed float64 `json:"gs,omitempty"`
	BaroRate    int16   `json:"baro_rate,omitempty"`
	Latitude    float64 `json:"lat,omitempty"`
	Longitude   float64 `json:"lon,omitempty"`
	NavHeading  float64 `json:"nav_heading,omitempty"`
	RSSi        float64 `json:"rssi,omitempty"`
	Category    string  `json:"category,omitempty"`
	Emergency   string  `json:"emergency,omitempty"`
	Messages    float64 `json:"messages,omitempty"`
}

// Define the aircraft metrics
var (
	dump1090AltBaro = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "alt_baro",
		Help:      "Barometric Altitude.",
	},
		[]string{"flight", "hex"},
	)
	dump1090AltGeom = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "alt_geom",
		Help:      "Geometric Altitude.",
	},
		[]string{"flight", "hex"},
	)
	dump1090BaroRate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "baro_rate",
		Help:      "Rate of Barometric Change.",
	},
		[]string{"flight", "hex"},
	)
	dump1090GroundSpeed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "gs",
		Help:      "Ground Speed.",
	},
		[]string{"flight", "hex"},
	)
	dump1090NavHeading = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "nav_heading",
		Help:      "Navigational Heading.",
	},
		[]string{"flight", "hex"},
	)
	dump1090Rssi = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "rssi",
		Help:      "Signal Strength.",
	},
		[]string{"flight", "hex"},
	)
	dump1090Messages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "messages",
		Help:      "Number of Messages.",
	},
		[]string{"flight", "hex"},
	)
)

func aircraftMetrics(aircraft []Aircraft) {

	dump1090AltBaro.Reset()
	dump1090AltGeom.Reset()
	dump1090BaroRate.Reset()
	dump1090GroundSpeed.Reset()
	dump1090NavHeading.Reset()
	dump1090Rssi.Reset()
	dump1090Messages.Reset()

	for _, s := range aircraft {
		labels := prometheus.Labels{"flight": strings.TrimSpace(s.Flight), "hex": s.Hex}
		dump1090AltBaro.With(labels).Set(float64(s.AltoBaro))
		dump1090AltGeom.With(labels).Set(float64(s.AltoGeom))
		dump1090BaroRate.With(labels).Set(float64(s.BaroRate))
		dump1090GroundSpeed.With(labels).Set(float64(s.GroundSpeed))
		dump1090NavHeading.With(labels).Set(float64(s.NavHeading))
		dump1090Rssi.With(labels).Set(float64(s.RSSi))
		dump1090Messages.With(labels).Set(float64(s.Messages))
	}

}

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

	aircraftMetrics(aircraft_list.Aircraft)
	// fmt.Printf("%+v\n", aircraft_list)
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
		readAircraftFile(path)
		// readStatsFile(path)
	}

}

func init() {
	// reg := prometheus.NewRegistry()
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(dump1090AltBaro)
	prometheus.MustRegister(dump1090AltGeom)
	prometheus.MustRegister(dump1090BaroRate)
	prometheus.MustRegister(dump1090GroundSpeed)
	prometheus.MustRegister(dump1090NavHeading)
	prometheus.MustRegister(dump1090Rssi)
	prometheus.MustRegister(dump1090Messages)
}

func main() {

	path := flag.String("path", "/run/dump1090-fa/", "Path to json files. Default /run/dump1090-fa/")
	port := flag.String("port", "3000", "Port to expose metrics")
	flag.Parse()
	fmt.Println("Path to json files:", *path)
	fmt.Println("Listen Port:", *port)

	go readFiles(*path)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+*port, nil))

}
