package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type GeoLocation struct {
	Status  string            `json:"status"`
	Results []GeocodingResult `json:"results"`
}

type GeocodingResult struct {
	AddressComponents []GeocodingAddressComponent `json:"address_components"`
}

type GeocodingAddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

/*
Extract a sane name from the geocoding object
*/
func (location *Location) Name() string {
	unknownLocation := "Unknown"
	var geoLocation GeoLocation
	err := json.Unmarshal([]byte(location.Geocoding), &geoLocation)
	if err != nil {
		log.Printf("Error decoding location object: %v", err)
		log.Printf("%v", location.Geocoding)
		return "Unknown"
	}

	if geoLocation.Status != "OK" || len(geoLocation.Results) == 0 {
		return unknownLocation
	}

	var postal_town, locality string

	for _, addresscomponents := range geoLocation.Results[0].AddressComponents {
		if stringSliceContains(addresscomponents.Types, "postal_town") {
			postal_town = addresscomponents.LongName
		}
		if stringSliceContains(addresscomponents.Types, "locality") {
			locality = addresscomponents.LongName
		}
	}
	if postal_town != "" {
		return postal_town
	}
	if locality != "" {
		return locality
	}
	return unknownLocation

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
