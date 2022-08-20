package main

import (
	"testing"
)

func TestDistance(t *testing.T) {
	var lat1 float64 = 50.72
	var lon1 float64 = -113.99
	var lat2 float64 = 51.72
	var lon2 float64 = -115.99

	distance := distance(lat1, lon1, lat2, lon2, "K")
	if distance != 178.21893191668582 {
		t.Errorf("Distance was incorrect")
	}
}
