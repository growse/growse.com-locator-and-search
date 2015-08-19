package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kpawlik/geojson"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"
)

type Location struct {
	Latitude             float64 `json:"lat" binding:"required"`
	Longitude            float64 `json:"long" binding:"required"`
	Geocoding            string
	Timestamp            time.Time
	DeviceTimestamp      time.Time
	DeviceTimestampAsInt int64   `json:"time" binding:"required"`
	Accuracy             float32 `json:"acc" binding:"required"`
	Distance             float64
	GSMType              string `json:"gsmtype" binding:"required"`
	WifiSSID             string `json:"wifissid" binding:"required"`
	DeviceID             string `json:"deviceid" binding:"required"`
}

type GeoLocation struct {
	Name string
}

type GeoName struct {
	Geonames []GeoLocation
}

func GetLastLoction() (*Location, error) {
	var location Location
	err := db.QueryRow("select geocoding, latitude,longitude,devicetimestamp from locations where geocoding ? 'geonames' order by devicetimestamp desc limit 1").Scan(&location.Geocoding, &location.Latitude, &location.Longitude, &location.Timestamp)

	return &location, err
}

func GetAverageSpeed() (float64, error) {
	var speed float64
	err := db.QueryRow("select 2.23693629*avg(speed) from (select distance/(extract(epoch from (devicetimestamp - lag(devicetimestamp) over (order by devicetimestamp asc)))::float) as speed from locations where extract(year from devicetimestamp at time zone 'UTC') = date_part('year', now() at time zone 'UTC')) a;").Scan(&speed)
	if err != nil {
		return 0, err
	}
	return speed, nil
}

func GetTotalDistance() (float64, error) {
	var distance float64
	err := db.QueryRow("select 0.000621371192*sum(distance) from locations where extract(year from devicetimestamp at time zone 'UTC') = date_part('year', now() at time zone 'UTC');").Scan(&distance)
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

/* HTTP handlers */

func WhereLineStringHandler(c *gin.Context) {
	linestring, err := GetLineStringAsJSON(c.Params.ByName("year"))
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
	}
	c.Data(200, "application/json", []byte(linestring))
}

func WhereHandler(c *gin.Context) {
	avgspeed, err := GetAverageSpeed()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	totaldistance, err := GetTotalDistance()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}

	lastlocation, err := GetLastLoction()
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
		return
	}
	obj := gin.H{"Title": "Where", "Stylesheet": stylesheetfilename, "Javascript": javascriptfilename, "Avgspeed": avgspeed, "Totaldistance": totaldistance, "LastLocation": lastlocation}
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	err = templates.ExecuteTemplate(buf, "where.html", obj)
	pageBytes := buf.Bytes()
	if err == nil {
		c.Data(200, "text/html", pageBytes)
	} else {
		InternalError(err)
		c.String(500, "Internal Error")
	}
}

func LocatorHandler(c *gin.Context) {
	locators := []Location{}

	err := c.Bind(&locators)
	if err != nil {
		c.String(400, fmt.Sprintf("%v", err))
		return
	}
	for _, locator := range locators {

		locator.DeviceTimestamp = time.Unix(locator.DeviceTimestampAsInt/1000, 1000000*(locator.DeviceTimestampAsInt%1000))
		locator.GetGeocoding()
		locator.GetRelativeSpeedDistance()
		log.Printf("Time: %v", locator.DeviceTimestamp)
		tx, err := db.Begin()
		if err != nil {
			InternalError(err)
			c.String(500, "Internal Error")
		}

		if locator.Geocoding != "" {
			_, err = tx.Exec("Update locations set geocoding=null")
			if err != nil {
				tx.Rollback()
				InternalError(err)
				c.String(500, "Internal Error")
				return
			}
		}
		_, err = tx.Exec("insert into locations (timestamp,devicetimestamp,latitude,longitude,accuracy,gsmtype,wifissid,geocoding,distance) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)", time.Now(), &locator.DeviceTimestamp, &locator.Latitude, &locator.Longitude, &locator.Accuracy, &locator.GSMType, &locator.WifiSSID, &locator.Geocoding, &locator.Distance)
		if err != nil {
			tx.Rollback()
			InternalError(err)
			c.String(500, "Database Error")
			return
		}

		err = tx.Commit()
		if err != nil {
			InternalError(err)
			c.String(500, "Database txn Error")
			return
		}
	}
	c.String(200, "Yay")
}

func (loc *Location) GetRelativeSpeedDistance() {
	prev := Location{}
	err := db.QueryRow("select devicetimestamp,latitude,longitude from locations order by devicetimestamp desc limit 1").Scan(&prev.DeviceTimestamp, &prev.Latitude, &prev.Longitude)
	if err != nil {
		log.Printf("Error found getting previous point. Setting distance to 0: %v", err)
		loc.Distance = 0
		return
	}
	if prev.Latitude == loc.Latitude && prev.Longitude == loc.Longitude {
		loc.Distance = 0
	} else {
		loc.Distance = 6378100 * DistanceOnUnitSphere(loc.Latitude, loc.Longitude, prev.Latitude, prev.Longitude)
	}
}

func (loc *Location) GetGeocoding() {
	if configuration.GeocodeApiURL == "" {
		InternalError(errors.New("Geocoding API should not be blank"))
		return
	}
	geocodingUrl := fmt.Sprintf(configuration.GeocodeApiURL, loc.Latitude, loc.Longitude)
	response, err := http.Get(geocodingUrl)
	if err != nil {
		InternalError(err)
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		InternalError(err)
		return
	}
	loc.Geocoding = string(body)
}

func DistanceOnUnitSphere(lat1 float64, long1 float64, lat2 float64, long2 float64) float64 {
	// Convert latitude and longitude to
	//spherical coordinates in radians.
	degrees_to_radians := math.Pi / 180.0

	// phi = 90 - latitude
	phi1 := (90.0 - lat1) * degrees_to_radians
	phi2 := (90.0 - lat2) * degrees_to_radians

	// theta = longitude
	theta1 := long1 * degrees_to_radians
	theta2 := long2 * degrees_to_radians

	// Compute spherical distance from spherical coordinates.

	// For two locations in spherical coordinates
	// (1, theta, phi) and (1, theta, phi)
	// cosine( arc length ) =
	// sin phi sin phi' cos(theta-theta') + cos phi cos phi'
	// distance = rho * arc length

	cos := (math.Sin(phi1)*math.Sin(phi2)*math.Cos(theta1-theta2) +
		math.Cos(phi1)*math.Cos(phi2))

	cos = math.Max(math.Min(cos, 1.0), -1.0)

	arc := math.Acos(cos)

	// Remember to multiply arc by the radius of the earth
	// in your favorite set of units to get length.
	return arc
}
