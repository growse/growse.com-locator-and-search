package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/martinlindhe/unit"
	"log"
	"net/http"
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
	DeviceID             string `json:"deviceid" binding:"required"`
	Altitude             float32
	VerticalAccuracy     float32
	Speed                float32
	Geocoding            string
}

func GetLastLocation() (*Location, error) {
	var location Location
	defer timeTrack(time.Now())
	query := "select " +
		"geocoding, " +
		"ST_Y(ST_AsText(point)), " +
		"ST_X(ST_AsText(point)) " +
		"from locations " +
		"where where geocoding is not null" +
		"order by devicetimestamp desc " +
		"limit 1"
	err := db.QueryRow(query).Scan(&location.Geocoding, &location.Latitude, &location.Longitude, &location.DeviceTimestamp)
	return &location, err
}

func GetTotalDistanceInMiles() (float64, error) {
	var distance float64
	defer timeTrack(time.Now())
	log.Println(time.Date(time.Now().UTC().Year(), 1, 1, 0, 0, 0, 0, time.UTC).String())
	err := db.QueryRow("select "+
		"sum(distance) "+
		"from "+
		"(select ST_Distance(point,lag(point,1,point) over (order by devicetimestamp asc)) as distance "+
		"from locations where devicetimestamp> $1"+
		") a;", time.Date(time.Now().UTC().Year(), 1, 1, 0, 0, 0, 0, time.UTC)).Scan(&distance)
	if err != nil {
		return 0, err
	}
	distanceInMeters := unit.Length(distance) * unit.Meter
	return distanceInMeters.Miles(), nil
}

func GetLocationsBetweenDates(from time.Time, to time.Time) (*[]Location, error) {
	defer timeTrack(time.Now())
	query := "select " +
		"coalesce(geocoding -> 'results' -> 0 ->> 'formatted_address', ''), " +
		"ST_Y(ST_AsText(point)), " +
		"ST_X(ST_AsText(point)), " +
		"devicetimestamp, " +
		"ST_Distance(point,lag(point,1,point) over (order by devicetimestamp asc)) as distance, " +
		"coalesce(speed, coalesce(3.6*ST_Distance(point,lag(point,1,point) over (order by devicetimestamp asc))/extract('epoch' from (devicetimestamp-lag(devicetimestamp) over (order by devicetimestamp asc))),0)) as speed, " +
		"coalesce(altitude, 0), " +
		"accuracy, " +
		"coalesce(verticalaccuracy, 0) " +
		"from locations where " +
		"devicetimestamp>$1 and devicetimestamp<$2 " +
		"order by devicetimestamp desc"
	rows, err := db.Query(query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var locations []Location
	for rows.Next() {
		var location Location
		err := rows.Scan(
			&location.Geocoding,
			&location.Latitude,
			&location.Longitude,
			&location.DeviceTimestamp,
			&location.Distance,
			&location.Speed,
			&location.Altitude,
			&location.Accuracy,
			&location.VerticalAccuracy,
		)
		if err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}
	return &locations, nil

}

/* HTTP handlers */
func LocationHandler(c *gin.Context) {
	location, err := GetLastLocation()
	distance, err := GetTotalDistanceInMiles()
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
	location, err := GetLastLocation()
	if err != nil {
		c.String(500, err.Error())
	}
	c.Header("Last-modified", location.Timestamp.Format("Mon, 02 Jal 2006 15:04:05 GMT"))
	c.Status(200)
}

func OTListUserHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"results": []string{"growse"},
	})
}

type OTPos struct {
	Tst  int64   `json:"tst" binding:"required"`
	Acc  float32 `json:"acc" binding:"required"`
	Type string  `json:"_type" binding:"required"`
	Alt  float32 `json:"alt" binding:"required"`
	Lon  float64 `json:"lon" binding:"required"`
	Vac  float32 `json:"vac" binding:"required"`
	Vel  float32 `json:"vel" binding:"required"`
	Lat  float64 `json:"lat" binding:"required"`
	Addr string  `json:"addr" binding:"required"`
}

func (location Location) toOT() OTPos {
	return OTPos{
		Tst:  location.DeviceTimestamp.Unix(),
		Acc:  location.Accuracy,
		Type: "location",
		Alt:  location.Altitude,
		Lat:  location.Latitude,
		Lon:  location.Longitude,
		Vel:  location.Speed,
		Vac:  location.VerticalAccuracy,
		Addr: location.Geocoding,
	}
}

func OTLastPosHandler(c *gin.Context) {
	location, err := GetLastLocation()
	if err != nil {
		c.String(500, err.Error())
		return
	}
	if location == nil {
		c.String(500, "No location found")
		return
	}
	last := location.toOT()
	c.JSON(200, []OTPos{
		last,
	})
}

func OTLocationsHandler(c *gin.Context) {
	const iso8061fmt = "2006-01-02T15:04:05"
	from := c.DefaultQuery("from", time.Now().AddDate(0, 0, -1).Format(iso8061fmt))
	to := c.DefaultQuery("to", time.Now().Format(iso8061fmt))
	fromTime, err := time.Parse(iso8061fmt, from)

	if err != nil {
		c.String(500, fmt.Sprintf("Invalid from time %v: %v", from, err))
		return
	}
	toTime, err := time.Parse(iso8061fmt, to)
	if err != nil {
		c.String(500, fmt.Sprintf("Invalid to time %v: %v", to, err))
		return
	}

	locations, err := GetLocationsBetweenDates(fromTime, toTime)
	if err != nil {
		c.String(500, err.Error())
		return
	}
	if locations == nil {
		c.String(500, "No locations found")
		return
	}
	var otpos []OTPos
	for _, location := range *locations {
		otpos = append(otpos, location.toOT())
	}
	response := struct {
		Data []OTPos `json:"data"`
	}{otpos}
	responseBytes, _ := json.Marshal(response)
	responseReader := bytes.NewReader(responseBytes)
	c.DataFromReader(200, int64(len(responseBytes)), "application/json", responseReader, nil)
}

func OTVersionHandler(c *gin.Context) {
	c.JSON(200, gin.H{"version": "1.0-growse-locator"})
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	// At the moment, this is just an echo impl. At some point publish new updates down this.
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v\n", err)
		return
	}

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		switch string(msg) {
		case "LAST":
			location, err := GetLastLocation()
			if err != nil {
				log.Printf("Error fetching last location: %v", err)
				break
			}
			locationAsBytes, err := json.Marshal(location.toOT())
			if err != nil {
				log.Printf("Error formatting location for websocket: %v", err)
				break
			}
			conn.WriteMessage(t, locationAsBytes)
			break
		default:
			break
		}
	}
}
