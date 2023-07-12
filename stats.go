package main

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
