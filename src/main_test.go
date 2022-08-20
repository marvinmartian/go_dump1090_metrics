package main

import (
	"fmt"
	"testing"
)

func TestDistance(t *testing.T) {
	var lat1 float64 = 50.72
	var lon1 float64 = -113.99
	var lat2 float64 = 51.72
	var lon2 float64 = -115.99

	distance := distance(lat1, lon1, lat2, lon2)
	if distance != 178.21893191668582 {
		t.Errorf("Distance was incorrect. Expected: 178.21893191668582")
	}
}

func TestRelativeAngle(t *testing.T) {
	var lat1 float64 = 50.72
	var lon1 float64 = -113.99
	var lat2 float64 = 51.72
	var lon2 float64 = -115.99

	rel_angle := relative_angle(lat1, lon1, lat2, lon2)
	fmt.Println(rel_angle)
	if rel_angle != 153.434948822922 {
		t.Errorf("Angle was incorrect. Expected: 153.434948822922")
	}
}
