package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kpawlik/geojson"
	"github.com/lib/pq"
	"log"
	"math"
	"strconv"
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

/*
Definately returns mph. Hence 2.236blahblah
*/
func GetAverageSpeed() (float64, error) {
	var speed float64
	err := db.QueryRow("select 2.23693629*avg(speed) from (select distance/(extract(epoch from (devicetimestamp - lag(devicetimestamp) over (order by devicetimestamp asc)))::float) as speed from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC')) = $1) a;", time.Now().UTC().Year()).Scan(&speed)
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
	err := db.QueryRow("select 0.000621371192*sum(distance) from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC')) = $1;", time.Now().UTC().Year()).Scan(&distance)
	if err != nil {
		return 0, err
	}
	return distance, nil
}

func GetLineStringAsJSON(year string, filtered bool) (string, error) {
	lineString := geojson.NewLineString(nil)
	var sqlStatement string
	if filtered {
		sqlStatement = "select kalmanlongitude,kalmanlatitude,kalmandistance from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC'))=$1 and kalmanaccuracy<(select percentile_disc(0.9) within group (order by kalmanaccuracy) from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC'))=$1) order by devicetimestamp asc"
	} else {
		sqlStatement = "select longitude,latitude,distance from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC'))=$1 and accuracy<(select percentile_disc(0.9) within group (order by accuracy) from locations where date_part('year'::text, date(devicetimestamp at time zone 'UTC'))=$1) order by devicetimestamp asc"
	}
	rows, err := db.Query(sqlStatement, year)
	if err != nil {
		return "", err
	}

	defer rows.Close()
	for rows.Next() {
		var coords geojson.Coordinate
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
	filtered, _ := strconv.ParseBool(c.Params.ByName("filtered"))
	linestring, err := GetLineStringAsJSON(c.Params.ByName("year"), filtered)
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
	}
	c.Data(200, "application/json", []byte(linestring))
}
func WhereLineStringHandlerNonFiltered(c *gin.Context) {
	linestring, err := GetLineStringAsJSON(c.Params.ByName("year"), false)
	if err != nil {
		InternalError(err)
		c.String(500, "Internal Error")
	}
	c.Data(200, "application/json", []byte(linestring))
}

func OSMWhereHandler(c *gin.Context) {
	filtered, err := strconv.ParseBool(c.Params.ByName("filtered"))
	if err != nil {
		filtered = false
	}
	year := c.Params.ByName("year")
	obj := gin.H{"Year": year, "Filtered": filtered}
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	err = templates.ExecuteTemplate(buf, "osm.html", obj)
	pageBytes := buf.Bytes()
	if err == nil {
		c.Data(200, "text/html", pageBytes)
	} else {
		InternalError(err)
		c.String(500, "Internal Error")
	}
}

func WhereHandler(c *gin.Context) {
	avgspeed, err := GetAverageSpeed()
	if err != nil {
		InternalError(err)
	}

	totaldistance, err := GetTotalDistance()
	if err != nil {
		InternalError(err)
	}

	lastlocation, err := GetLastLoction()
	if err != nil {
		InternalError(err)
	}
	obj := gin.H{
		"Title":           "Where",
		"Stylesheet":      stylesheetfilename,
		"Javascript":      javascriptfilename,
		"WhereJavascript": wherejavascriptfilename,
		"Avgspeed":        avgspeed,
		"Totaldistance":   totaldistance,
		"LastLocation":    lastlocation}
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

/*
Receive POST from phone. This should be an application/json containing an array of points.
*/

func LocatorHandler(c *gin.Context) {
	c.String(204, "Deprecated")
	return
	locators := []Location{}

	err := c.BindJSON(&locators)
	if err != nil {
		c.String(400, fmt.Sprintf("%v", err))
		return
	}

	newLocation := false
	for _, locator := range locators {
		locator.DeviceTimestamp = time.Unix(locator.DeviceTimestampAsInt / 1000, 1000000 * (locator.DeviceTimestampAsInt % 1000))
		locator.GetRelativeSpeedDistance(db)

		_, err = db.Exec("insert into locations (timestamp,devicetimestamp,latitude,longitude,accuracy,gsmtype,wifissid,distance) values ($1,$2,$3,$4,$5,$6,$7,$8)", time.Now(), &locator.DeviceTimestamp, &locator.Latitude, &locator.Longitude, &locator.Accuracy, &locator.GSMType, &locator.WifiSSID, &locator.Distance)

		switch i := err.(type) {
		case nil:
			newLocation = true
			continue
		case *pq.Error:
			log.Printf("Managed to get a duplicate timestamp: %v", locator)
		default:
			log.Printf("%T", err)
			InternalError(i)
			c.String(500, "Database Error")
			return
		}

	}
	//Now to update the geocoding from the latest locator
	if newLocation {
		GeocodingWorkQueue <- true
	}
	c.String(200, "Yay")
}

func (loc *Location) GetRelativeSpeedDistance(thisDb *sql.DB) {
	prev := Location{}
	err := thisDb.QueryRow("select devicetimestamp,latitude,longitude from locations where devicetimestamp<$1 order by devicetimestamp desc limit 1", loc.DeviceTimestamp).Scan(&prev.DeviceTimestamp, &prev.Latitude, &prev.Longitude)
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

	cos := (math.Sin(phi1) * math.Sin(phi2) * math.Cos(theta1 - theta2) +
		math.Cos(phi1) * math.Cos(phi2))

	cos = math.Max(math.Min(cos, 1.0), -1.0)

	arc := math.Acos(cos)

	// Remember to multiply arc by the radius of the earth
	// in your favorite set of units to get length.
	return arc
}
