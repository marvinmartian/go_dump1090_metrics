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
	Messages float64    `json:"messages,int"`
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
		Name:      "messages_total",
		Help:      "Number of Messages.",
	},
		[]string{"time_period"},
	)
	dump1090Distance = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "distance",
		Help:      "Distance from receiver.",
	},
		[]string{"flight", "hex"},
	)
	dump1090MaxRangeDirection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "recent_aircraft_max_range_by_direction",
		Help:      "Max distance by direction.",
	},
		[]string{"direction", "time_period"},
	)
	dump1090MaxRange = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "recent_aircraft_max_range",
		Help:      "Maximum range of recently observed aircraft.",
	},
		[]string{"time_period"},
	)
	dump1090CountByDirection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "recent_aircraft_with_direction",
		Help:      "Aircraft count by direction.",
	},
		[]string{"direction", "time_period"},
	)
	dump1090CountWithPos = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "recent_aircraft_with_position",
		Help:      "Number of aircraft with position.",
	},
		[]string{"time_period"},
	)
	dump1090CountWithMlat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "dump1090",
		Name:      "recent_aircraft_with_multilateration",
		Help:      "Number of aircraft with multilateration position.",
	},
		[]string{"time_period"},
	)
	// dump1090Observed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Namespace: "dump1090",
	// 	Name:      "recent_aircraft_observed",
	// 	Help:      "Number of aircraft with multilateration position.",
	// },
	// 	[]string{"time_period"},
	// )
)

var opsMetrics struct {
	AirCraftFileReads func() prometheus.Counter `name:"aircraft_file_reads" help:"Number of reads on the aircraft file"`
	StatsFileReads    func() prometheus.Counter `name:"stats_file_reads" help:"Number of reads on the stats file"`
}

var metrics struct {
	// MessagesTotal func(statLabels) prometheus.Gauge `name:"messages_total" help:"Total number of messages received"`

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

	StatsLocalAccepted       func(statLabels) prometheus.Gauge `name:"stats_local_accepted" help:"Number of valid Mode S messages accepted with N-bit errors corrected"`
	StatsLocalSignalStrength func(statLabels) prometheus.Gauge `name:"stats_local_signal_strength_dbFS" help:"Signal strength dbFS"`
	StatsLocalStrongSignal   func(statLabels) prometheus.Gauge `name:"stats_local_strong_signals" help:"Number of messages that had a signal power above -3dBFS"`
	StatsLocalPeakSignal     func(statLabels) prometheus.Gauge `name:"stats_local_peak_signal_strength_dbFS" help:"Peak signal strength dbFS"`

	// dump1090_stats_local_accepted{time_period="last1min"} 786
	// dump1090_stats_local_bad{time_period="last1min"} 3889369
	// dump1090_stats_local_modeac{time_period="last1min"} 0
	// dump1090_stats_local_modes{time_period="last1min"} 1652142
	// dump1090_stats_local_noise_level_dbFS{time_period="last1min"} -20.4
	// dump1090_stats_local_peak_signal_strength_dbFS{time_period="last1min"} -2.5
	// dump1090_stats_local_samples_dropped{time_period="last1min"} 0
	// dump1090_stats_local_samples_processed{time_period="last1min"} 144048128
	// dump1090_stats_local_signal_strength_dbFS{time_period="last1min"} -7.7
	// dump1090_stats_local_strong_signals{time_period="last1min"} 6
	// dump1090_stats_local_unknown_icao{time_period="last1min"} 675782
}

type statLabels struct {
	TimePeriod string `label:"time_period"`
}

type requestLabels struct {
	Flight string `label:"flight"`
	Hex    string `label:"hex"`
}

const radius = 6371.0e3

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

var direction_lut = [16]string{"N", "NE", "NE", "E", "E", "SE", "SE", "S", "S", "SW", "SW", "W", "W", "NW", "NW", "N"}

func degrees2radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func relative_angle(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {

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

func relative_direction(angle float64) string {
	index := int(math.Abs(angle) / 22.5)
	// fmt.Println(index)
	return direction_lut[index]
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
	for d := range direction_lut {
		aircraft_direction[direction_lut[d]] = 0
		aircraft_direction_max_range[direction_lut[d]] = 0
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
					angle := relative_angle(ReceiverLat, ReceiverLon, s.Latitude, s.Longitude)
					direction := relative_direction(angle)
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
	// fmt.Println(aircraft_observed)
	// fmt.Println(aircraft_with_mlat)
	// fmt.Println(aircraft_with_pos)
	// fmt.Println(aircraft_direction_max_range)

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
			metrics.StatsLocalAccepted(minuteLabel).Set(value.Local.Accepted[0])
			metrics.StatsLocalSignalStrength(minuteLabel).Set(value.Local.SignalStrength)
			metrics.StatsLocalStrongSignal(minuteLabel).Set(value.Local.StrongSignals)
			metrics.StatsLocalPeakSignal(minuteLabel).Set(value.Local.PeakSignal)
			// dump1090Messages.With(prometheus.Labels{"time_period": key}).Set(float64(value.Messages))
		}

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
	var aircraft_list AircraftList

	// Unmarshal to aircraft list
	json.Unmarshal(byteValue, &aircraft_list)

	aircraftMetrics(aircraft_list)
	// fmt.Printf("%+v\n", aircraft_list)
}

func readStatsFile(path string) {

	// Open the file
	jsonFile, err := os.Open(path + "stats.json")

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

	// go readAircraftFile(path, aircraftTicker)
	// go readStatsFile(path, statsTicker)

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
