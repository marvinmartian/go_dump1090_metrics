package main

import "github.com/prometheus/client_golang/prometheus"

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
)

var opsMetrics struct {
	AirCraftFileReads func() prometheus.Counter `name:"aircraft_file_reads" help:"Number of reads on the aircraft file"`
	StatsFileReads    func() prometheus.Counter `name:"stats_file_reads" help:"Number of reads on the stats file"`
}

var metrics struct {
	MessagesTotal func(statLabels) prometheus.Gauge `name:"stats_messages_total" help:"Total number of Mode-S messages processed"`

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

	LocalAccepted         func(statLabels) prometheus.Gauge `name:"stats_local_accepted" help:"Number of valid Mode S messages accepted with N-bit errors corrected"`
	LocalSignalStrength   func(statLabels) prometheus.Gauge `name:"stats_local_signal_strength_dbFS" help:"Signal strength dbFS"`
	LocalStrongSignal     func(statLabels) prometheus.Gauge `name:"stats_local_strong_signals" help:"Number of messages that had a signal power above -3dBFS"`
	LocalPeakSignal       func(statLabels) prometheus.Gauge `name:"stats_local_peak_signal_strength_dbFS" help:"Peak signal strength dbFS"`
	LocalBad              func(statLabels) prometheus.Gauge `name:"stats_local_bad" help:"Number of Mode S preambles that didn't result in a valid message"`
	LocalModeAc           func(statLabels) prometheus.Gauge `name:"stats_local_modeac" help:"Mode A/C preambles decoded"`
	LocalModes            func(statLabels) prometheus.Gauge `name:"stats_local_modes" help:"Number of Mode S preambles received"`
	LocalNoiseLevel       func(statLabels) prometheus.Gauge `name:"stats_local_noise_level_dbFS" help:"Noise level dbFS"`
	LocalSamplesDropped   func(statLabels) prometheus.Gauge `name:"stats_local_samples_dropped" help:"Number of samples dropped"`
	LocalSamplesProcessed func(statLabels) prometheus.Gauge `name:"stats_local_samples_processed" help:"Number of samples processed"`
	LocalUnknownIcao      func(statLabels) prometheus.Gauge `name:"stats_local_unknown_icao" help:"Number of Mode S preambles containing unrecognized ICAO"`

	// from here
	// StatsCprGlobalSkipped func(statLabels) prometheus.Gauge `name:"stats_cpr_global_skipped" help:"Global position attempts skipped due to missing data"`
	// StatsCprGlobalSpeed   func(statLabels) prometheus.Gauge `name:"stats_cpr_global_speed" help:"Global positions rejected due to speed check"`
	// StatsCprLocalSpeed func(statLabels) prometheus.Gauge `name:"stats_cpr_local_speed" help:"Local positions rejected due to speed check"`

	RemoteAccepted    func(statLabels) prometheus.Gauge `name:"stats_remote_accepted" help:"Number of valid Mode S messages accepted with N-bit errors corrected"`
	RemoteBad         func(statLabels) prometheus.Gauge `name:"stats_remote_bad" help:"Number of Mode S preambles that didn't result in a valid message"`
	RemoteModeAc      func(statLabels) prometheus.Gauge `name:"stats_remote_modeac" help:"Number of Mode A/C preambles decoded"`
	RemoteModes       func(statLabels) prometheus.Gauge `name:"stats_remote_modes" help:"Number of Mode S preambles received"`
	RemoteUnknownIcao func(statLabels) prometheus.Gauge `name:"stats_remote_unknown_icao" help:"Number of Mode S preambles containing unrecognized ICAO"`

	TracksAll           func(statLabels) prometheus.Gauge `name:"stats_tracks_all" help:"Number of tracks created"`
	TracksSingleMessage func(statLabels) prometheus.Gauge `name:"stats_tracks_single_message" help:"Number of tracks consisting of only a single message"`
}

type statLabels struct {
	TimePeriod string `label:"time_period"`
}

type requestLabels struct {
	Flight string `label:"flight"`
	Hex    string `label:"hex"`
}
