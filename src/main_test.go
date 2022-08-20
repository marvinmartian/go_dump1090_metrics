package main

import "testing"

func TestDistance(t *testing.T) {
	distance := distance(foo, foo)
	if distance != 10 {
		t.Errorf("Distance was incorrect")
	}
}
