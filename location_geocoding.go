package main

import (
	"encoding/json"
	"log"
	"errors"
	"fmt"
	"net/http"
	"io/ioutil"
	"time"
)

/*
Extract a sane name from the geocoding object
*/
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

func (loc *Location) GetGeocoding() string {
	if configuration.GeocodeApiURL == "" {
		InternalError(errors.New("Geocoding API should not be blank"))
		return ""
	}
	geocodingUrl := fmt.Sprintf(configuration.GeocodeApiURL, loc.Latitude, loc.Longitude)
	start := time.Now()
	response, err := http.Get(geocodingUrl)
	duration := time.Since(start)
	log.Printf("Reverse geocoded in %v", duration)
	if err != nil {
		log.Printf("Error getting geolocation from API: %v", err)
		return ""
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		InternalError(errors.New(fmt.Sprintf("invalid response from Geolocation API: %v %v", response.StatusCode, response.Body)))
		return ""
	}
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		InternalError(err)
		return ""
	}
	return string(body)
}

func UpdateLatestLocationWithGeocoding(workChan <-chan bool) {
	log.Print("Starting geocoding goroutine")
	for {
		_, more := <-workChan
		if more {
			log.Print("Updating latest geocoding")
			var location Location
			var id int
			err := db.QueryRow("select id,latitude,longitude from locations order by devicetimestamp desc limit 1").Scan(&id, &location.Latitude, &location.Longitude)
			if err != nil {
				InternalError(err)
			}
			tx, err := db.Begin()
			if err != nil {
				InternalError(err)
			}
			geocoding := location.GetGeocoding()
			if geocoding != "" {
				_, err = tx.Exec("Update locations set geocoding=null")
				if err != nil {
					tx.Rollback()
					InternalError(err)
				}
				_, err = tx.Exec("update locations set geocoding=$1 where id=$2", geocoding, id)
				if err != nil {
					log.Printf("Location that caused fail is: %s", geocoding)
					tx.Rollback()
					InternalError(err)
				}
				err = tx.Commit()
				if err != nil {
					InternalError(err)
				} else {
					log.Print("Geocoding complete")
				}
			}
		} else {
			log.Print("Got signal, quitting geocoding goroutine.")
			return
		}

	}
}