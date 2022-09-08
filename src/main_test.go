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

func TestDegrees2Radians(t *testing.T) {

	var tests = []struct {
		degrees float64
		want    float64
	}{
		{234.2376, 4.08821735196947},
		{186.3456, 3.2523442666043456},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%f", tt.degrees)
		t.Run(testname, func(t *testing.T) {
			ans := degrees2radians(tt.degrees)
			if ans != tt.want {
				t.Errorf("got %f, want %f", ans, tt.want)
			}
		})
	}

}

func TestContains(t *testing.T) {
	var tests = []struct {
		haystack []string
		needle   string
		want     bool
	}{
		{[]string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}, "SE", true},
		{[]string{"foo", "bar"}, "foobar", false},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s,%s", tt.haystack, tt.needle)
		t.Run(testname, func(t *testing.T) {
			ans := contains(tt.haystack, tt.needle)
			if ans != tt.want {
				t.Errorf("got %t, want %t", ans, tt.want)
			}
		})
	}

}
