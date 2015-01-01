package main

import (
	"encoding/json"
	"github.com/kpawlik/geojson"
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
	err := db.QueryRow("select geocoding, latitude,longitude,timestamp from locations where geocoding like '%geonames%' order by timestamp desc limit 1").Scan(&location.Geocoding, &location.Latitude, &location.Longitude, &location.Timestamp)

	return &location, err
}

func GetAverageSpeed() (float64, error) {
	var speed float64
	err := db.QueryRow("select 2.23693629*avg(distance/(timedelta/1000000::float)) from locations where extract(year from devicetimestamp at time zone 'UTC') = date_part('year', now() at time zone 'UTC')-1;").Scan(&speed)
	if err != nil {
		return 0, err
	}
	return speed, nil
}

func GetTotalDistance() (float64, error) {
	var distance float64
	err := db.QueryRow("select 0.000621371192*sum(distance) from locations where extract(year from devicetimestamp at time zone 'UTC') = date_part('year', now() at time zone 'UTC')-1;").Scan(&distance)
	if err != nil {
		return 0, err
	}
	return distance, nil
}

func GetLineStringAsJSON(year string) (string, error) {

	lineString := geojson.NewLineString(nil)
	rows, err := db.Query("select longitude,latitude from locations where extract (year from devicetimestamp at time zone 'UTC')=$1 order by devicetimestamp asc", year)
	if err != nil {
		return "", err
	}

	defer rows.Close()
	for rows.Next() {
		var coords geojson.Coordinate
		rows.Scan(&coords[0], &coords[1])
		lineString.AddCoordinates(coords)
	}
	feature := geojson.NewFeature(lineString, nil, nil)
	featureCollection := geojson.NewFeatureCollection([]*geojson.Feature{feature})
	json, err := geojson.Marshal(featureCollection)
	if err != nil {
		return "", err
	}
	return json, nil
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
