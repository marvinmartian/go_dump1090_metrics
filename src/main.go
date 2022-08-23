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
	Messages float64    `json:"messages"`
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
	Start    float64    `json:"start"`
	End      float64    `json:"end"`
	Messages float64    `json:"messages"`
	Local    StatLocal  `json:"local"`
	Cpr      StatCpr    `json:"cpr"`
	Cpu      StatCpu    `json:"cpu"`
	Remote   StatRemote `json:"remote"`
	Track    StatTrack  `json:"tracks"`
}

type StatLocal struct {
	Accepted         []float64 `json:"accepted"`
	Bad              float64   `json:"bad"`
	ModeAc           float64   `json:"modeac"`
	Modes            float64   `json:"modes"`
	Noise            float64   `json:"noise"`
	PeakSignal       float64   `json:"peak_signal"`
	SamplesDropped   float64   `json:"samples_dropped"`
	SamplesProcessed float64   `json:"samples_processed"`
	SignalStrength   float64   `json:"signal"`
	StrongSignals    float64   `json:"strong_signals"`
	UnknownIcao      float64   `json:"unknown_icao"`
}

type StatRemote struct {
	Accepted    []float64 `json:"accepted"`
	Bad         float64   `json:"bad"`
	ModeAc      float64   `json:"modeac"`
	Modes       float64   `json:"modes"`
	UnknownIcao float64   `json:"unknown_icao"`
}

type StatCpu struct {
	Demod      float64 `json:"demod"`
	Reader     float64 `json:"reader"`
	Background float64 `json:"background"`
}

type StatTrack struct {
	All           float64 `json:"all"`
	SingleMessage float64 `json:"single_message"`
}

type StatCpr struct {
	Airborne      float64 `json:"airborne"`
	Filtered      float64 `json:"filtered"`
	GlobalBad     float64 `json:"global_bad"`
	GlobalOk      float64 `json:"global_ok"`
	GlobalRange   float64 `json:"global_range"`
	GlobalSkipped float64 `json:"global_skipped"`
	GlobalSpeed   float64 `json:"global_speed"`

	LocalOk               float64 `json:"local_ok"`
	LocalAircraftRelative float64 `json:"local_aircraft_relative"`
	LocalReceiverRelative float64 `json:"local_receiver_relative"`
	LocalSkipped          float64 `json:"local_skipped"`
	LocalRange            float64 `json:"local_range"`
	LocalSpeed            float64 `json:"local_speed"`

	Surface float64 `json:"surface"`
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
	MessagesTotal func(statLabels) prometheus.Gauge `name:"messages_total" help:"Total number of messages received"`

	RecentAircraftObserved func(statLabels) prometheus.Gauge `name:"recent_aircraft_observed" help:"Recent Aircraft observed"`

	CprAirborne      func(statLabels) prometheus.Gauge `name:"stats_cpr_airborne" help:"cpr airborne"`
	CprFiltered      func(statLabels) prometheus.Gauge `name:"stats_cpr_filtered" help:"cpr fltered"`
	CprGlobalBad     func(statLabels) prometheus.Gauge `name:"stats_cpr_global_bad" help:"cpr global bad"`
	CprGlobalOk      func(statLabels) prometheus.Gauge `name:"stats_cpr_global_ok" help:"cpr global ok"`
	CprGlobalRange   func(statLabels) prometheus.Gauge `name:"stats_cpr_global_range" help:"cpr global range"`
	CprGlobalSkipped func(statLabels) prometheus.Gauge `name:"stats_cpr_global_skipped" help:"cpr global skipped"`
	CprGlobalSpeed   func(statLabels) prometheus.Gauge `name:"stats_cpr_global_speed" help:"cpr global speed"`

	CprLocalAircraftRelative func(statLabels) prometheus.Gauge `name:"stats_cpr_local_aircraft_relative" help:"cpr local aircraft relative"`
	CprLocalOk               func(statLabels) prometheus.Gauge `name:"stats_cpr_local_ok" help:"cpr local ok"`
	CprLocalRange            func(statLabels) prometheus.Gauge `name:"stats_cpr_local_range" help:"cpr local range"`
	CprLocalReceiverRelative func(statLabels) prometheus.Gauge `name:"stats_cpr_local_receiver_relative" help:"cpr local receiver relative"`
	CprLocalSkipped          func(statLabels) prometheus.Gauge `name:"stats_cpr_local_skipped" help:"cpr local skipped"`
	CprLocalSpeed            func(statLabels) prometheus.Gauge `name:"stats_cpr_local_speed" help:"cpr local speed"`
	CprSurface               func(statLabels) prometheus.Gauge `name:"stats_cpr_surface" help:"cpr surface"`

	CpuBackgroundMs func(statLabels) prometheus.Gauge `name:"stats_cpu_background_milliseconds" help:"background cpu"`
	CpuDemodMs      func(statLabels) prometheus.Gauge `name:"stats_cpu_demod_milliseconds" help:"Demod ms"`
	CpuReaderMs     func(statLabels) prometheus.Gauge `name:"stats_cpu_reader_milliseconds" help:"Reader ms"`
}

type statLabels struct {
	TimePeriod string `label:"time_period"`
}

type requestLabels struct {
	Flight string `label:"flight"`
	Hex    string `label:"hex"`
}

const radius = 6371

var direction_lut = [16]string{"N", "NE", "NE", "E", "E", "SE", "SE", "S", "S", "SW", "SW", "W", "W", "NW", "NW", "N"}

func degrees2radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func relative_angle(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	angle := math.Atan2(lat2-lat1, lng2-lng1) * (180 / math.Pi)
	return angle
}

func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	degreesLat := degrees2radians(lat2 - lat1)
	degreesLong := degrees2radians(lng2 - lng1)
	a := (math.Sin(degreesLat/2)*math.Sin(degreesLat/2) +
		math.Cos(degrees2radians(lat1))*
			math.Cos(degrees2radians(lat2))*math.Sin(degreesLong/2)*
			math.Sin(degreesLong/2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := radius * c
	return d
}

func relative_direction(angle float64) string {
	// direction_lut := [16]string{"N", "NE", "NE", "E", "E", "SE", "SE", "S", "S", "SW", "SW", "W", "W", "NW", "NW", "N"}
	index := int(angle/22.5) + 1
	return direction_lut[index]
}

func aircraftMetrics(aircraft []Aircraft) {

	dump1090AltBaro.Reset()
	dump1090Distance.Reset()
	dump1090AltGeom.Reset()
	dump1090BaroRate.Reset()
	dump1090GroundSpeed.Reset()
	dump1090NavHeading.Reset()
	dump1090Rssi.Reset()
	// dump1090Messages.Reset()

	for _, s := range aircraft {
		labels := prometheus.Labels{"flight": strings.TrimSpace(s.Flight), "hex": s.Hex}
		if s.Latitude != 0 {
			dist := distance(ReceiverLat, ReceiverLon, s.Latitude, s.Longitude)
			// angle := relative_angle(ReceiverLat, ReceiverLon, s.Latitude, s.Longitude)
			// direction := relative_direction(angle)
			// fmt.Println(angle)
			// fmt.Println(direction)
			dump1090Distance.With(labels).Set(dist)
		}

		dump1090AltBaro.With(labels).Set(float64(s.AltoBaro))
		dump1090AltGeom.With(labels).Set(float64(s.AltoGeom))
		dump1090BaroRate.With(labels).Set(float64(s.BaroRate))
		dump1090GroundSpeed.With(labels).Set(float64(s.GroundSpeed))
		dump1090NavHeading.With(labels).Set(float64(s.NavHeading))
		dump1090Rssi.With(labels).Set(float64(s.RSSi))
		// dump1090Messages.With(labels).Set(float64(s.Messages))
	}

}

func statMetrics(stats Statistics) {

	m := make(map[string]SingleStat)
	// m["last5minute"] = stats.Last_5
	m["last1min"] = stats.Last_1
	m["latest"] = stats.Latest

	for key, value := range m {
		minuteLabel := statLabels{
			TimePeriod: key,
		}

		metrics.MessagesTotal(minuteLabel).Set(value.Messages)

		metrics.CprAirborne(minuteLabel).Set(value.Cpr.Airborne)
		metrics.CprFiltered(minuteLabel).Set(value.Cpr.Filtered)
		metrics.CprGlobalBad(minuteLabel).Set(value.Cpr.GlobalBad)
		metrics.CprGlobalOk(minuteLabel).Set(value.Cpr.GlobalOk)
		metrics.CprGlobalRange(minuteLabel).Set(value.Cpr.GlobalRange)

		metrics.CprLocalAircraftRelative(minuteLabel).Set(value.Cpr.LocalAircraftRelative)
		metrics.CprLocalOk(minuteLabel).Set(value.Cpr.LocalOk)
		metrics.CprLocalRange(minuteLabel).Set(value.Cpr.LocalRange)
		metrics.CprLocalReceiverRelative(minuteLabel).Set(value.Cpr.LocalReceiverRelative)
		metrics.CprLocalSkipped(minuteLabel).Set(value.Cpr.LocalSkipped)

		metrics.CprSurface(minuteLabel).Set(value.Cpr.Surface)

		metrics.CpuBackgroundMs(minuteLabel).Set(value.Cpu.Background)
		metrics.CpuDemodMs(minuteLabel).Set(value.Cpu.Demod)
		metrics.CpuReaderMs(minuteLabel).Set(value.Cpu.Reader)
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

	statMetrics(stats)

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

}

func readFilesTicker(path string) {

	aircraftTicker := time.NewTicker(5 * time.Second)
	statsTicker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			<-aircraftTicker.C
			readAircraftFile(path)
		}
	}()

	go func() {
		for {
			<-statsTicker.C
			readStatsFile(path)
		}
	}()

	// go readAircraftFile(path, aircraftTicker)
	// go readStatsFile(path, statsTicker)

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

	go readFilesTicker(*path)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+*port, nil))

}
