package main

import (
	"testing"
)

func TestDistance(t *testing.T) {
	var lat1 float64 = 50.72
	var lon1 float64 = -113.99
	var lat2 float64 = 51.72
	var lon2 float64 = -115.99

	distance := distance(lat1, lon1, lat2, lon2)
	if distance != 178218.93191668583 {
		t.Errorf("Distance was incorrect. Expected: 178.21893191668582")
	}
}

func TestRelativeAngle(t *testing.T) {
	var lat1 float64 = 50.72
	var lon1 float64 = -113.99
	var lat2 float64 = 51.72
	var lon2 float64 = -115.99

	rel_angle := relative_angle(lat1, lon1, lat2, lon2)
	if rel_angle != 296.565051177078 {
		t.Errorf("Angle was incorrect. Expected: 296.565051177078")
	}
}

func TestRelativeDirection(t *testing.T) {
	var angle float64 = 116.4048736325465

	rel_angle := relative_direction(angle)
	if rel_angle != "SE" {
		t.Errorf("Direction was incorrect. Expected: SE")
	}
}
