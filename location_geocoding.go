package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

	var postalTown, locality string

	for _, addresscomponents := range geoLocation.Results[0].AddressComponents {
		if stringSliceContains(addresscomponents.Types, "postal_town") {
			postalTown = addresscomponents.LongName
		}
		if stringSliceContains(addresscomponents.Types, "locality") {
			locality = addresscomponents.LongName
		}
	}
	if postalTown != "" {
		return postalTown
	}
	if locality != "" {
		return locality
	}
	return unknownLocation

}

func GetGeocoding(place string) (string, error) {
	if configuration.GeocodeApiURL == "" {
		err := errors.New("Geocoding API should not be blank")
		InternalError(err)
		return "", err
	}
	if place == "" {
		err := errors.New("Place should not be blank")
		InternalError(err)
		return "", err
	}
	geocodingUrl := fmt.Sprintf(configuration.GeocodeApiURL, url.QueryEscape(place))
	return fetchGeocodingResponse(geocodingUrl)
}

func (location *Location) GetReverseGeocoding() (string, error) {
	if configuration.ReverseGeocodeApiURL == "" {
		err := errors.New("Reverse Geocoding API should not be blank")
		InternalError(err)
		return "", err
	}
	geocodingUrl := fmt.Sprintf(configuration.ReverseGeocodeApiURL, location.Latitude, location.Longitude)
	return fetchGeocodingResponse(geocodingUrl)
}

func fetchGeocodingResponse(geocodingUrl string) (string, error) {
	defer timeTrack(time.Now())
	response, err := http.Get(geocodingUrl)

	if err != nil {
		log.Printf("Error getting geolocation from API: %v", err)
		return "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		if err == nil {
			body = []byte("")
		}
		err := errors.New(fmt.Sprintf("invalid response from Geolocation API: %v %v", response.StatusCode, body))
		InternalError(err)
		return "", err
	}

	if err != nil {
		InternalError(err)
		return "", err
	}
	return string(body), nil
}

func UpdateLatestLocationWithGeocoding(workChan <-chan bool) {
	log.Print("Starting geocoding goroutine")
	for {
		_, more := <-workChan
		if more {
			log.Print("Updating latest geocoding")
			var location Location
			var id int
			err := db.QueryRow("select id,ST_Y(ST_AsText(point)),ST_X(ST_AsText(point)) from locations order by devicetimestamp desc limit 1").Scan(&id, &location.Latitude, &location.Longitude)
			if err != nil {
				InternalError(err)
			}
			tx, err := db.Begin()
			if err != nil {
				InternalError(err)
			}
			geocoding, err := location.GetReverseGeocoding()
			if err != nil {
				InternalError(err)
			} else {
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

func GeocodingCrawler(quitChan <-chan bool) {
	log.Print("Starting geocoding backlog crawler")
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			var location Location
			var id int
			err := db.QueryRow("select id,ST_Y(ST_AsText(point)),ST_X(ST_AsText(point)) from locations where geocoding is null and devicetimestamp<CURRENT_DATE - 1 order by devicetimestamp desc limit 1").Scan(&id, &location.Latitude, &location.Longitude)
			if err != nil {
				log.Print("Error fetching latest location without geocode")
				break
			}
			geocoding, err := location.GetReverseGeocoding()
			if err != nil {
				log.Printf("Error reversing geocode for: %v", location)
				break
			} else {
				_, err = db.Exec("update locations set geocoding=$1 where id=$2", geocoding, id)
				if err != nil {
					log.Printf("Location that caused fail is: %s", geocoding)
				} else {
					log.Printf("Geocoded location id=%v", id)
				}
			}
		case <-quitChan:

			log.Print("Clos")
			return
		}
	}
}
