package main

import (
	"encoding/json"
	"log"
	"time"
)

type Location struct {
	Latitude        float32
	Longitude       float32
	Geocoding       string
	Timestamp       time.Time
	DeviceTimestamp time.Time
	Accuracy        float32
	TimeDelta       int64
	Distance        float32
	GSMType         string
	WifiSSID        string
}

type GeoLocation struct {
	/*Distance    string
	CountryId   string*/
	Name string
	/*CountryCode string
	GeonameId   int
	ToponymName string
	Fcode       string
	FclName     string
	FcodeName   string
	CountryName string
	Lat         string
	Long        string
	AdminName1  string
	Fcl         string
	AdminCode1  string
	Population  int*/
}

type GeoName struct {
	Geonames []GeoLocation
}

func GetLastLoction() (*Location, error) {
	var location Location
	err := db.QueryRow("select geocoding, latitude,longitude from locations where geocoding like '%geonames%' order by timestamp desc limit 1").Scan(&location.Geocoding, &location.Latitude, &location.Longitude)

	return &location, err
}

func (location *Location) Name() string {
	var msg GeoName
	err := json.Unmarshal([]byte(location.Geocoding), &msg)
	if err != nil {
		log.Printf("Error decoding location object: %v", err)
		return "Unknown"
	}
	if len(msg.Geonames) > 0 {
		return msg.Geonames[0].Name
	}
	log.Printf("No geonames found in: %v", location.Geocoding)
	return "Unknown"
}
