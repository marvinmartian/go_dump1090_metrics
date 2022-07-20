// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type AircraftList struct {
	Messages int32      `json:"messages"`
	Aircraft []Aircraft `json:"aircraft"`
}

type Aircraft struct {
	Hex      string `json:"hex"`
	Flight   string `json:"flight,omitempty"`
	AltoBaro int16  `json:"alt_baro,omitempty"`
}

func readFile(file string) {

	// Open the file
	jsonFile, err := os.Open(file)

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

	fmt.Printf("%+v\n", aircraft_list)
}

func main() {

	file := flag.String("file", "/run/dump1090-fa/aircraft.json", "Path to aircraft.json file")
	flag.Parse()
	fmt.Println(*file)

	ticker := time.NewTicker(5 * time.Second)

	for _ = range ticker.C {
		fmt.Println("Ticking..")
		readFile(*file)
	}

}
