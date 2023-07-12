package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/hashicorp/go-memdb"
)

type AircraftDetails struct {
	Icao24           string `csv:"icao24"`
	Registration     string `csv:"registration"`
	ManufacturerIcao string `csv:"manufacturericao"`
	ManufacturerName string `csv:"manufacturername"`
	Model            string `csv:"model"`
	Operator         string `csv:"operator"`
}

var db *memdb.MemDB

func downloadFile(filepath string, url string) (err error) {
	fmt.Println("downloading file")
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		// log.Fatal("failed to create file")
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		// log.Fatal("error in http get file")
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		// log.Fatal("error in writing file to disk")
		return err
	}

	return nil
}

func parseCsv(path string, db *memdb.MemDB) {

	aircraftFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer aircraftFile.Close()

	txn := db.Txn(true)

	aircraft := []*AircraftDetails{}

	fmt.Println(path)
	if err := gocsv.UnmarshalFile(aircraftFile, &aircraft); err != nil { // Load clients from file
		panic(err)
	}
	// aircraft2 := []*AircraftDetails{
	// 	{"391927", "Piper", "foo", "bar", "banana"},
	// 	{"3d3191", "G-115", "foo", "bar", "banana"},
	// }
	// for _, a := range aircraft2 {
	// 	if err := txn.Insert("aircraft", a); err != nil {
	// 		panic(err)
	// 	}
	// }

	for _, aircraft := range aircraft {
		if aircraft.Icao24 != "" {
			// fmt.Println("Hello", aircraft.Icao24)
			// fmt.Println("Icao", aircraft.Icao24)
			// fmt.Println("Model", aircraft.Model)
			// fmt.Println("ManufacturerIcao", aircraft.ManufacturerIcao)
			// fmt.Println("ManufacturerName", aircraft.ManufacturerName)
			// if aircraft.Status != "" {
			// 	fmt.Println("Status", aircraft.Status)
			// }

			if aircraft.Registration != "" {
				if aircraft.Registration[0:1] == "C" {
					// fmt.Println(firstCharacter)
					if err := txn.Insert("aircraft", aircraft); err != nil {
						panic(err)
					}
				}

			}

		}

	}
	txn.Commit()

	// txn = db.Txn(false)
	// defer txn.Abort()

	// raw, err2 := txn.First("aircraft", "id", "c04101")
	// if err2 != nil {
	// 	panic(err2)
	// }

	// // Say hi!
	// fmt.Printf("Hello %s!\n", raw.(*AircraftDetails))

}

func dbSetup() *memdb.MemDB {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"aircraft": {
				Name: "aircraft",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Icao24"},
					},
					"registration": {
						Name:         "registration",
						Unique:       false,
						AllowMissing: true,
						Indexer:      &memdb.StringFieldIndex{Field: "Registration"},
					},
					"manufacturerIcao": {
						Name:         "manufacturerIcao",
						Unique:       false,
						AllowMissing: true,
						Indexer:      &memdb.StringFieldIndex{Field: "ManufacturerIcao"},
					},
					"manufacturerName": {
						Name:         "manufacturerName",
						Unique:       false,
						AllowMissing: true,
						Indexer:      &memdb.StringFieldIndex{Field: "ManufacturerName"},
					},
					"model": {
						Name:         "model",
						Unique:       false,
						AllowMissing: true,
						Indexer:      &memdb.StringFieldIndex{Field: "Model"},
					},
				},
			},
		},
	}

	// Create a new data base
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}
	return db
}

func FindAircraft(icao string) *AircraftDetails {
	// fmt.Println(db)
	txn := db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("aircraft", "id", icao)
	if err != nil {
		panic("Error in db lookup")
	}
	// fmt.Printf("Hello %s!\n", raw.(*AircraftDetails))
	return raw.(*AircraftDetails)
}

func flightInit() {
	var aircraftDbUrl string = "https://opensky-network.org/datasets/metadata/aircraftDatabase.csv"
	// var zipFilePath string = "/home/melvin/Projects/go_dump1090_metrics/aircraftDatabase.zip"
	var csvFilePath string = "./aircraftDatabase.csv"

	// Setup BadgerDB
	// opt := badger.DefaultOptions("").WithInMemory(true)
	// db, err := badger.Open(opt)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// fmt.Println(aircraftDbUrl)
	info, err := os.Stat(csvFilePath)
	if err != nil {
		// TODO: handle errors (e.g. file not found)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("doesnt exist")
			downloadFile(csvFilePath, aircraftDbUrl)
		}
	} else {
		if info.ModTime().Before(time.Now().AddDate(0, 0, -7)) {
			downloadFile(csvFilePath, aircraftDbUrl)
		}
	}

	db = dbSetup()
	parseCsv(csvFilePath, db)

}
