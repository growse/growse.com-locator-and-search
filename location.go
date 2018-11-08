package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/kpawlik/geojson"
	"time"
)

/*
This should be some sort of thing that's sent from the phone
*/
type Location struct {
	Latitude             float64 `json:"lat" binding:"required"`
	Longitude            float64 `json:"long" binding:"required"`
	Timestamp            time.Time
	DeviceTimestamp      time.Time
	DeviceTimestampAsInt int64   `json:"time" binding:"required"`
	Accuracy             float32 `json:"acc" binding:"required"`
	Distance             float64
	GSMType              string `json:"gsmtype" binding:"required"`
	WifiSSID             string `json:"wifissid" binding:"required"`
	DeviceID             string `json:"deviceid" binding:"required"`
	Geocoding            string
}

func GetLastLoction() (*Location, error) {
	var location Location
	defer timeTrack(time.Now())
	err := db.QueryRow("select geocoding, latitude,longitude,devicetimestamp from locations where geocoding is not null order by devicetimestamp desc limit 1").Scan(&location.Geocoding, &location.Latitude, &location.Longitude, &location.Timestamp)
	return &location, err
}

/*
Definitely returns mph. Hence 2.236blahblah
*/
func GetAverageSpeed() (float64, error) {
	var speed float64
	defer timeTrack(time.Now())
	err := db.QueryRow("select 2.23693629*sum(gisdistance)/extract(epoch from now()-date_trunc('year',now())) from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC')) = $1", time.Now().UTC().Year()).Scan(&speed)
	if err != nil {
		return 0, err
	}
	return speed, nil
}

/*
In miles.
*/
func GetTotalDistance() (float64, error) {
	var distance float64
	defer timeTrack(time.Now())
	err := db.QueryRow("select 0.000621371192*sum(gisdistance) from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC')) = $1;", time.Now().UTC().Year()).Scan(&distance)
	if err != nil {
		return 0, err
	}
	return distance, nil
}

func GetLineStringAsJSON(year string) (string, error) {
	lineString := geojson.NewLineString(nil)
	sqlStatement := "select longitude,latitude,gisdistance from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC'))=$1 and accuracy<(select percentile_disc(0.9) within group (order by accuracy) from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC'))=$1) order by devicetimestamp asc"
	rows, err := db.Query(sqlStatement, year)
	if err != nil {
		return "", err
	}

	defer rows.Close()
	for rows.Next() {
		var coords geojson.Coordinate
		coords = geojson.Coordinate{0, 0}
		var distance float32
		rows.Scan(&coords[0], &coords[1], &distance)
		if distance > 100 {
			// We only want to add points where something's actually moved significantly, this is in metres
			lineString.AddCoordinates(coords)
		}
	}
	// Dump the stuff into some sort of geojson thingie
	feature := geojson.NewFeature(lineString, nil, nil)
	featureCollection := geojson.NewFeatureCollection([]*geojson.Feature{feature})
	json, err := geojson.Marshal(featureCollection)
	if err != nil {
		return "", err
	}
	return json, nil
}

/* HTTP handlers */
func WhereLineStringHandler(c *gin.Context) {
	linestring, err := GetLineStringAsJSON(c.Params.ByName("year"))
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error\n"+err.Error())
	}
	c.Data(200, "application/json", []byte(linestring))
}

/*
Receive POST from phone. This should be an application/json containing an array of points.
*/

func LocatorHandler(c *gin.Context) {
	c.String(204, "Deprecated")
	return
}

func LocationHandler(c *gin.Context) {
	location, err := GetLastLoction()
	distance, err := GetTotalDistance()
	if err != nil {
		c.String(500, err.Error())
	}
	c.Header("Last-modified", location.Timestamp.Format("Mon, 02 Jal 2006 15:04:05 GMT"))
	c.JSON(200, gin.H{
		"name":          location.Name(),
		"latitude":      fmt.Sprintf("%.2f", location.Latitude),
		"longitude":     fmt.Sprintf("%.2f", location.Longitude),
		"totalDistance": humanize.FormatFloat("#,###.##", distance),
	})
}

func LocationHeadHandler(c *gin.Context) {
	location, err := GetLastLoction()
	if err != nil {
		c.String(500, err.Error())
	}
	c.Header("Last-modified", location.Timestamp.Format("Mon, 02 Jal 2006 15:04:05 GMT"))
	c.Status(200)
}
