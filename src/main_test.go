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

	rel_angle := relativeAngle(lat1, lon1, lat2, lon2)
	if rel_angle != 296.565051177078 {
		t.Errorf("Angle was incorrect. Expected: 296.565051177078")
	}
}

func TestRelativeDirection(t *testing.T) {
	var acceptedDirection = []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	for i := 0; i <= 360; i++ {
		direction := relativeDirection(float64(i))
		if contains(acceptedDirection, direction) == false {
			t.Errorf("Direction was incorrect.")
		}
	}

}
