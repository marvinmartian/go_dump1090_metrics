// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cabify/gotoprom"
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
	Start    float32    `json:"start"`
	End      float32    `json:"end"`
	Messages int32      `json:"messages"`
	Local    StatLocal  `json:"local"`
	Cpr      StatCpr    `json:"cpr"`
	Cpu      StatCpu    `json:"cpu"`
	Remote   StatRemote `json:"remote"`
	Track    StatTrack  `json:"tracks"`
}

type StatLocal struct {
	Accepted         []float32 `json:"accepted"`
	Bad              float32   `json:"bad"`
	ModeAc           float32   `json:"modeac"`
	Modes            float32   `json:"modes"`
	Noise            float32   `json:"noise"`
	PeakSignal       float32   `json:"peak_signal"`
	SamplesDropped   float32   `json:"samples_dropped"`
	SamplesProcessed float32   `json:"samples_processed"`
	SignalStrength   float32   `json:"signal"`
	StrongSignals    float32   `json:"strong_signals"`
	UnknownIcao      float32   `json:"unknown_icao"`
}

type StatRemote struct {
	Accepted    []float32 `json:"accepted"`
	Bad         float32   `json:"bad"`
	ModeAc      float32   `json:"modeac"`
	Modes       float32   `json:"modes"`
	UnknownIcao float32   `json:"unknown_icao"`
}

type StatCpu struct {
	Demod      float32 `json:"demod"`
	Reader     float32 `json:"reader"`
	Background float32 `json:"background"`
}

type StatTrack struct {
	All           float32 `json:"all"`
	SingleMessage float32 `json:"single_message"`
}

type StatCpr struct {
	Airborne      float32 `json:"airborne"`
	Filtered      float32 `json:"filtered"`
	GlobalBad     float32 `json:"global_bad"`
	GlobalOk      float32 `json:"global_ok"`
	GlobalRange   float32 `json:"global_range"`
	GlobalSkipped float32 `json:"global_skipped"`
	GlobalSpeed   float32 `json:"global_speed"`

	LocalOk               float32 `json:"local_ok"`
	LocalAircraftRelative float32 `json:"local_aircraft_relative"`
	LocalReceiverRelative float32 `json:"local_receiver_relative"`
	LocalSkipped          float32 `json:"local_skipped"`
	LocalRange            float32 `json:"local_range"`
	LocalSpeed            float32 `json:"local_speed"`

	Surface float32 `json:"surface"`
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
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

var (
	ReceiverLat float64
	ReceiverLon float64
)

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
	dump1090Distance = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "distance",
		Help:      "Distance from receiver.",
	},
		[]string{"flight", "hex"},
	)
)

var metrics struct {
	// SomeCounter                      func() prometheus.Counter   `name:"some_counter" help:"some counter"`
	// SomeHistogram                    func() prometheus.Histogram `name:"some_histogram" help:"Some histogram with default prometheus buckets" buckets:""`
	// SomeHistogramWithSpecificBuckets func() prometheus.Histogram `name:"some_histogram_with_buckets" help:"Some histogram with custom buckets" buckets:".01,.05,.1"`
	SomeGauge func(requestLabels) prometheus.Gauge `name:"some_gauge" help:"Some gauge"`
	// SomeSummaryWithSpecificMaxAge    func() prometheus.Summary   `name:"some_summary_with_specific_max_age" help:"Some summary with custom max age" max_age:"20m" objectives:"0.50,0.95,0.99"`

	// Requests struct {
	// 	Total func(requestLabels) prometheus.Count `name:"total" help:"Total amount of requests served"`
	// } `namespace:"requests"`
}

type requestLabels struct {
	Flight string `label:"flight"`
	Hex    string `label:"hex"`
}

func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64, unit ...string) float64 {
	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(math.Pi * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515

	if len(unit) > 0 {
		if unit[0] == "K" {
			dist = dist * 1.609344
		} else if unit[0] == "N" {
			dist = dist * 0.8684
		}
	}

	return dist
}

func aircraftMetrics(aircraft []Aircraft) {

	dump1090AltBaro.Reset()
	dump1090Distance.Reset()
	dump1090AltGeom.Reset()
	dump1090BaroRate.Reset()
	dump1090GroundSpeed.Reset()
	dump1090NavHeading.Reset()
	dump1090Rssi.Reset()
	dump1090Messages.Reset()

	for _, s := range aircraft {
		labels := prometheus.Labels{"flight": strings.TrimSpace(s.Flight), "hex": s.Hex}
		if s.Latitude != 0 {
			dist := distance(ReceiverLat, ReceiverLon, s.Latitude, s.Longitude, "K")
			dump1090Distance.With(labels).Set(dist)
		}

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

	metrics.SomeGauge(requestLabels{Flight: "Foo", Hex: "Bar"}).Set(100)

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
	fmt.Println("-------------")
}

func readReceiverInfo(path string) {

	// Open the file
	jsonFile, err := os.Open(path + "receiver.json")

	// Print the error if that happens.
	if err != nil {
		fmt.Println(err)
	}

	// defer file close
	defer jsonFile.Close()

	// read file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// // Initialize list of coordinates
	var cords Coordinate

	// // Unmarshal to aircraft list
	json.Unmarshal(byteValue, &cords)

	ReceiverLat = cords.Lat
	ReceiverLon = cords.Lon

	// fmt.Printf("%+v\n", cords)

}

func readFiles(path string) {
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		readAircraftFile(path)
		readStatsFile(path)
	}

}

func init() {
	// reg := prometheus.NewRegistry()
	gotoprom.MustInit(&metrics, "dump1090")
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(dump1090AltBaro)
	prometheus.MustRegister(dump1090AltGeom)
	prometheus.MustRegister(dump1090BaroRate)
	prometheus.MustRegister(dump1090GroundSpeed)
	prometheus.MustRegister(dump1090NavHeading)
	prometheus.MustRegister(dump1090Rssi)
	prometheus.MustRegister(dump1090Messages)
	prometheus.MustRegister(dump1090Distance)
}

func main() {

	path := flag.String("path", "/run/dump1090-fa/", "Path to json files. Default /run/dump1090-fa/")
	port := flag.String("port", "3000", "Port to expose metrics")
	flag.Parse()
	fmt.Println("Path to json files:", *path)
	fmt.Println("Listen Port:", *port)

	readReceiverInfo(*path)
	go readFiles(*path)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+*port, nil))

}
