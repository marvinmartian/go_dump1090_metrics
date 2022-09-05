// main.go
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cabify/gotoprom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AircraftList struct {
	Messages float64    `json:"messages,int"`
	Aircraft []Aircraft `json:"aircraft"`
}

type Aircraft struct {
	Hex         string   `json:"hex"`
	Flight      string   `json:"flight,omitempty"`
	AltoBaro    uint16   `json:"alt_baro,omitempty"`
	AltoGeom    uint16   `json:"alt_geom,omitempty"`
	GroundSpeed float64  `json:"gs,omitempty"`
	BaroRate    int16    `json:"baro_rate,omitempty"`
	Latitude    float64  `json:"lat,omitempty"`
	Longitude   float64  `json:"lon,omitempty"`
	NavHeading  float64  `json:"nav_heading,omitempty"`
	RSSi        float64  `json:"rssi,omitempty"`
	Category    string   `json:"category,omitempty"`
	Emergency   string   `json:"emergency,omitempty"`
	Messages    float64  `json:"messages,omitempty"`
	Seen        float64  `json:"seen,omitempty"`
	SeenPos     float64  `json:"seen_pos,omitempty"`
	Mlat        []string `json:"mlat,omitempty"`
}
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

var (
	ReceiverLat float64
	ReceiverLon float64
)

const radius = 6371.0e3

var myClient = &http.Client{Timeout: 10 * time.Second}
var directionLut = [17]string{"N", "NE", "NE", "E", "E", "SE", "SE", "S", "S", "SW", "SW", "W", "W", "NW", "NW", "N"}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func degrees2radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func relativeAngle(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {

	deg := math.Atan((lng2-lng1)/(lat2-lat1)) * (180 / math.Pi)

	if lat2 == lat1 {
		if lng2 > lng1 {
			return 90
		} else {
			return 270
		}
	}

	if lat2 > lat1 {
		return math.Mod(360+deg, 360)
	} else {
		return 180 + deg
	}
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

func relativeDirection(angle float64) string {
	if angle >= 360 {
		angle = 0
	}
	index := int(math.Abs(angle) / 22.5)
	return directionLut[index]
}

func aircraftMetrics(aircraftList AircraftList) {

	aircraft := aircraftList.Aircraft

	dump1090AltBaro.Reset()
	dump1090Distance.Reset()
	dump1090AltGeom.Reset()
	dump1090BaroRate.Reset()
	dump1090GroundSpeed.Reset()
	dump1090NavHeading.Reset()
	dump1090Rssi.Reset()
	dump1090MaxRangeDirection.Reset()
	dump1090MaxRange.Reset()
	// dump1090Messages.Reset()

	var threshold float64 = 15
	var aircraft_observed int = 0
	var aircraft_with_mlat int = 0
	var aircraft_with_pos float64 = 0
	var aircraft_max_range float64 = 0

	aircraft_direction := make(map[string]int)
	aircraft_direction_max_range := make(map[string]float64)
	for d := range directionLut {
		aircraft_direction[directionLut[d]] = 0
		aircraft_direction_max_range[directionLut[d]] = 0
	}
	// fmt.Println(aircraft_direction)
	// fmt.Println(aircraft_direction_max_range)
	// fmt.Println(threshold)

	// fmt.Println(float64(int(aircraftList.Messages)))
	// metrics.MessagesTotal(statLabels{TimePeriod: "latest"}).Set(aircraftList.Messages)

	for _, s := range aircraft {

		labels := prometheus.Labels{"flight": strings.TrimSpace(s.Flight), "hex": s.Hex}
		if s.Seen < threshold {
			aircraft_observed++
			if s.SeenPos < threshold {
				aircraft_with_pos++
				if contains(s.Mlat, "lat") {
					aircraft_with_mlat++
				}
				if s.Latitude != 0 {
					dist := distance(ReceiverLat, ReceiverLon, s.Latitude, s.Longitude)
					angle := relativeAngle(ReceiverLat, ReceiverLon, s.Latitude, s.Longitude)
					direction := relativeDirection(angle)
					// fmt.Println(angle)
					// fmt.Println(direction)
					aircraft_direction[direction]++
					dump1090CountByDirection.With(prometheus.Labels{"direction": direction, "time_period": "latest"}).Set(float64(aircraft_direction[direction]))
					if dist > float64(aircraft_direction_max_range[direction]) {
						aircraft_direction_max_range[direction] = dist
						dump1090MaxRangeDirection.With(prometheus.Labels{"direction": direction, "time_period": "latest"}).Set(dist)
					}
					if dist > aircraft_max_range {
						// Set Max Range Metric
						aircraft_max_range = dist
						dump1090MaxRange.With(prometheus.Labels{"time_period": "latest"}).Set(dist)
					}
					dump1090Distance.With(labels).Set(dist)
				}
			}

		}

		// details := FindAircraft(s.Hex)
		// fmt.Println(details)

		dump1090AltBaro.With(labels).Set(float64(s.AltoBaro))
		dump1090AltGeom.With(labels).Set(float64(s.AltoGeom))
		dump1090BaroRate.With(labels).Set(float64(s.BaroRate))
		dump1090GroundSpeed.With(labels).Set(float64(s.GroundSpeed))
		dump1090NavHeading.With(labels).Set(float64(s.NavHeading))
		dump1090Rssi.With(labels).Set(float64(s.RSSi))

	}

	metrics.RecentAircraftObserved(statLabels{TimePeriod: "latest"}).Set(float64(aircraft_observed))
	// fmt.Println(aircraft_direction)
	// dump1090Observed.With(prometheus.Labels{"time_period": "latest"}).Set(float64(aircraft_observed))
	dump1090Messages.With(prometheus.Labels{"time_period": "latest"}).Set(aircraftList.Messages)
	dump1090CountWithMlat.With(prometheus.Labels{"time_period": "latest"}).Set(float64(aircraft_with_mlat))
	dump1090CountWithPos.With(prometheus.Labels{"time_period": "latest"}).Set(float64(aircraft_with_pos))

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

		if key != "latest" {
			// metrics.MessagesTotal(minuteLabel).Set(value.Messages)
			dump1090Messages.With(prometheus.Labels{"time_period": key}).Set(value.Messages)
			metrics.LocalAccepted(minuteLabel).Set(value.Local.Accepted[0])
			metrics.LocalSignalStrength(minuteLabel).Set(value.Local.SignalStrength)
			metrics.LocalStrongSignal(minuteLabel).Set(value.Local.StrongSignals)
			metrics.LocalPeakSignal(minuteLabel).Set(value.Local.PeakSignal)
			metrics.LocalBad(minuteLabel).Set(value.Local.Bad)
			metrics.LocalModeAc(minuteLabel).Set(value.Local.ModeAc)
			metrics.LocalModes(minuteLabel).Set(value.Local.Modes)
			metrics.LocalNoiseLevel(minuteLabel).Set(value.Local.Noise)
			metrics.LocalSamplesDropped(minuteLabel).Set(value.Local.SamplesDropped)
			metrics.LocalSamplesProcessed(minuteLabel).Set(value.Local.SamplesProcessed)
			metrics.LocalUnknownIcao(minuteLabel).Set(value.Local.UnknownIcao)

			metrics.CprLocalSpeed(minuteLabel).Set(value.Cpr.LocalSpeed)
			metrics.CprGlobalSpeed(minuteLabel).Set(value.Cpr.GlobalSpeed)

			metrics.RemoteAccepted(minuteLabel).Set(value.Remote.Accepted[0])
			metrics.RemoteBad(minuteLabel).Set(value.Remote.Bad)
			metrics.RemoteModeAc(minuteLabel).Set(value.Remote.ModeAc)
			metrics.RemoteModes(minuteLabel).Set(value.Remote.Modes)
			metrics.RemoteUnknownIcao(minuteLabel).Set(value.Remote.UnknownIcao)

			metrics.TracksAll(minuteLabel).Set(value.Track.All)
			metrics.TracksSingleMessage(minuteLabel).Set(value.Track.SingleMessage)

			metrics.MessagesTotal(minuteLabel).Set(value.Messages)
			// dump1090Messages.With(prometheus.Labels{"time_period": key}).Set(float64(value.Messages))
		}

		metrics.CprAirborne(minuteLabel).Set(value.Cpr.Airborne)
		metrics.CprFiltered(minuteLabel).Set(value.Cpr.Filtered)
		metrics.CprGlobalBad(minuteLabel).Set(value.Cpr.GlobalBad)
		metrics.CprGlobalOk(minuteLabel).Set(value.Cpr.GlobalOk)
		metrics.CprGlobalRange(minuteLabel).Set(value.Cpr.GlobalRange)
		metrics.CprGlobalSkipped(minuteLabel).Set(value.Cpr.GlobalSkipped)

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

	var aircraft_path string = path + "aircraft.json"
	_, err := url.ParseRequestURI(aircraft_path)
	if err != nil {
		if _, err := os.Stat(aircraft_path); errors.Is(err, os.ErrNotExist) {
			// Do something because it doesn't exist
		} else {
			// Open the file
			jsonFile, err := os.Open(aircraft_path)

			// Increment the prom read metric
			opsMetrics.AirCraftFileReads().Inc()

			// Print the error if that happens.
			if err != nil {
				fmt.Println(err)
			}

			// defer file close
			defer jsonFile.Close()

			// read file
			byteValue, _ := ioutil.ReadAll(jsonFile)

			// Initialize list of aircraft
			var aircraftList AircraftList

			// Unmarshal to aircraft list
			json.Unmarshal(byteValue, &aircraftList)

			aircraftMetrics(aircraftList)
		}
	} else {
		aircraftList := new(AircraftList)
		getJson(aircraft_path, aircraftList)
		aircraftMetrics(*aircraftList)
	}
}

func readStatsFile(path string) {

	var stats_path string = path + "stats.json"
	_, err := url.ParseRequestURI(stats_path)
	if err != nil {
		if _, err := os.Stat(stats_path); errors.Is(err, os.ErrNotExist) {
			// Do something because it doesn't exist
		} else {
			// Open the file
			jsonFile, err := os.Open(stats_path)

			// Increment the prom read metric
			opsMetrics.StatsFileReads().Inc()

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
	} else {
		stats := new(Statistics)
		getJson(stats_path, stats)
		statMetrics(*stats)
	}

}

func readReceiverInfo(path string) {

	var receiver_path string = path + "receiver.json"
	_, err := url.ParseRequestURI(receiver_path)
	if err != nil {
		// fmt.Println("it's a filepath")
		if _, err := os.Stat(receiver_path); errors.Is(err, os.ErrNotExist) {
			// Do something because it doesn't exist
		} else {
			// Open the file
			jsonFile, err := os.Open(receiver_path)

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
	} else {
		cords := new(Coordinate)
		getJson(receiver_path, cords)
		ReceiverLat = cords.Lat
		ReceiverLon = cords.Lon
	}

}

func readFilesTicker(path string) {

	aircraftTicker := time.NewTicker(5 * time.Second)
	statsTicker := time.NewTicker(30 * time.Second)

	readAircraftFile(path)
	readStatsFile(path)

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

}

func init() {
	// reg := prometheus.NewRegistry()
	gotoprom.MustInit(&metrics, "dump1090")

	gotoprom.MustInit(&opsMetrics, "dump1090")

	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(dump1090AltBaro)
	prometheus.MustRegister(dump1090AltGeom)
	prometheus.MustRegister(dump1090BaroRate)
	prometheus.MustRegister(dump1090GroundSpeed)
	prometheus.MustRegister(dump1090NavHeading)
	prometheus.MustRegister(dump1090Rssi)
	prometheus.MustRegister(dump1090Messages)
	prometheus.MustRegister(dump1090Distance)
	prometheus.MustRegister(dump1090MaxRangeDirection)
	prometheus.MustRegister(dump1090MaxRange)
	prometheus.MustRegister(dump1090CountByDirection)
	prometheus.MustRegister(dump1090CountWithPos)
	prometheus.MustRegister(dump1090CountWithMlat)
	// prometheus.MustRegister(dump1090Observed)

}

func main() {

	path := flag.String("path", "/run/dump1090-fa/", "Path to json files. Default /run/dump1090-fa/")
	port := flag.String("port", "3000", "Port to expose metrics")
	flag.Parse()
	fmt.Println("Path to json files:", *path)
	fmt.Println("Listen Port:", *port)

	readReceiverInfo(*path)

	go readFilesTicker(*path)

	// go flightInit()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+*port, nil))

}
