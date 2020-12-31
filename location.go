package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	if db == nil {
		return nil, errors.New("No database connection available")
	}
	var location Location
	defer timeTrack(time.Now())
	query := "select " +
		"geocoding, " +
		"ST_Y(ST_AsText(point)), " +
		"ST_X(ST_AsText(point)) " +
		"from locations " +
		"where geocoding is not null " +
		"order by devicetimestamp desc " +
		"limit 1"
	err := db.QueryRow(query).Scan(&location.Geocoding, &location.Latitude, &location.Longitude)
	return &location, err
}

func GetTotalDistanceInMiles() (float64, error) {
	if db == nil {
		return 0, errors.New("No database connection available")
	}
	var distance float64
	defer timeTrack(time.Now())
	err := db.QueryRow("select distance from locations_distance_this_year").Scan(&distance)
	if err != nil {
		return 0, err
	}
	distanceInMeters := unit.Length(distance) * unit.Meter
	return distanceInMeters.Miles(), nil
}

func GetLocationsBetweenDates(from time.Time, to time.Time) (*[]Location, error) {
	if db == nil {
		return nil, errors.New("No database connection available")
	}
	defer timeTrack(time.Now())
	query := "select " +
		"coalesce(geocoding -> 'results' -> 0 ->> 'formatted_address', ''), " +
		"ST_Y(ST_AsText(point)), " +
		"ST_X(ST_AsText(point)), " +
		"devicetimestamp, " +
		"coalesce(speed, coalesce(3.6*ST_Distance(point,lag(point,1,point) over (order by devicetimestamp asc))/extract('epoch' from (devicetimestamp-lag(devicetimestamp) over (order by devicetimestamp asc))),0)) as speed, " +
		"coalesce(altitude, 0), " +
		"accuracy, " +
		"coalesce(verticalaccuracy, 0) " +
		"from locations where " +
		"devicetimestamp>=$1 and devicetimestamp<$2 " +
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
		return
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
		return
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

type GeocodingResponse struct {
	Documentation string `json:"documentation"`
	Licenses      []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"licenses"`
	Rate struct {
		Limit     int `json:"limit"`
		Remaining int `json:"remaining"`
		Reset     int `json:"reset"`
	} `json:"rate"`
	Results []struct {
		Annotations struct {
			DMS struct {
				Lat string `json:"lat"`
				Lng string `json:"lng"`
			} `json:"DMS"`
			MGRS       string `json:"MGRS"`
			Maidenhead string `json:"Maidenhead"`
			Mercator   struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"Mercator"`
			OSM struct {
				EditURL string `json:"edit_url"`
				NoteURL string `json:"note_url"`
				URL     string `json:"url"`
			} `json:"OSM"`
			UNM49 struct {
				Regions struct {
					EUROPE         string `json:"EUROPE"`
					PT             string `json:"PT"`
					SOUTHERNEUROPE string `json:"SOUTHERN_EUROPE"`
					WORLD          string `json:"WORLD"`
				} `json:"regions"`
				StatisticalGroupings []string `json:"statistical_groupings"`
			} `json:"UN_M49"`
			Callingcode int `json:"callingcode"`
			Currency    struct {
				AlternateSymbols     []interface{} `json:"alternate_symbols"`
				DecimalMark          string        `json:"decimal_mark"`
				HTMLEntity           string        `json:"html_entity"`
				IsoCode              string        `json:"iso_code"`
				IsoNumeric           string        `json:"iso_numeric"`
				Name                 string        `json:"name"`
				SmallestDenomination int           `json:"smallest_denomination"`
				Subunit              string        `json:"subunit"`
				SubunitToUnit        int           `json:"subunit_to_unit"`
				Symbol               string        `json:"symbol"`
				SymbolFirst          int           `json:"symbol_first"`
				ThousandsSeparator   string        `json:"thousands_separator"`
			} `json:"currency"`
			Flag     string  `json:"flag"`
			Geohash  string  `json:"geohash"`
			Qibla    float64 `json:"qibla"`
			Roadinfo struct {
				DriveOn string `json:"drive_on"`
				SpeedIn string `json:"speed_in"`
			} `json:"roadinfo"`
			Sun struct {
				Rise struct {
					Apparent     int `json:"apparent"`
					Astronomical int `json:"astronomical"`
					Civil        int `json:"civil"`
					Nautical     int `json:"nautical"`
				} `json:"rise"`
				Set struct {
					Apparent     int `json:"apparent"`
					Astronomical int `json:"astronomical"`
					Civil        int `json:"civil"`
					Nautical     int `json:"nautical"`
				} `json:"set"`
			} `json:"sun"`
			Timezone struct {
				Name         string `json:"name"`
				NowInDst     int    `json:"now_in_dst"`
				OffsetSec    int    `json:"offset_sec"`
				OffsetString string `json:"offset_string"`
				ShortName    string `json:"short_name"`
			} `json:"timezone"`
			What3Words struct {
				Words string `json:"words"`
			} `json:"what3words"`
			Wikidata string `json:"wikidata"`
		} `json:"annotations"`
		Bounds struct {
			Northeast struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"northeast"`
			Southwest struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"southwest"`
		} `json:"bounds"`
		Components struct {
			ISO31661Alpha2 string `json:"ISO_3166-1_alpha-2"`
			ISO31661Alpha3 string `json:"ISO_3166-1_alpha-3"`
			Category       string `json:"_category"`
			Type           string `json:"_type"`
			City           string `json:"city"`
			Continent      string `json:"continent"`
			Country        string `json:"country"`
			CountryCode    string `json:"country_code"`
			County         string `json:"county"`
			CountyCode     string `json:"county_code"`
			PoliticalUnion string `json:"political_union"`
			State          string `json:"state"`
			StateDistrict  string `json:"state_district"`
		} `json:"components"`
		Confidence int    `json:"confidence"`
		Formatted  string `json:"formatted"`
		Geometry   struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
	} `json:"results"`
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
	StayInformed struct {
		Blog    string `json:"blog"`
		Twitter string `json:"twitter"`
	} `json:"stay_informed"`
	Thanks    string `json:"thanks"`
	Timestamp struct {
		CreatedHTTP string `json:"created_http"`
		CreatedUnix int    `json:"created_unix"`
	} `json:"timestamp"`
	TotalResults int `json:"total_results"`
}

func PlaceHandler(c *gin.Context) {
	if db == nil {
		c.String(500, "No database connection available")
		c.Abort()
		return
	}
	place := c.PostForm("place")
	geocodingResponse, err := GetGeocoding(place)
	if err != nil {
		InternalError(err)
		c.String(500, err.Error())
		c.Abort()
		return
	}
	var geocoding GeocodingResponse
	err = json.Unmarshal([]byte(geocodingResponse), &geocoding)
	if err != nil {
		InternalError(err)
		c.String(500, err.Error())
		c.Abort()
		return
	}
	if len(geocoding.Results) == 1 {
		geometry := geocoding.Results[0].Geometry
		rows, err := db.Query("select avg(st_distance(POINT,ST_SetSRID(ST_MakePoint($1, $2),4326)))/1000 as distance,date(devicetimestamp) from locations group by date(devicetimestamp) order by distance asc limit 10;", geometry.Lng, geometry.Lat)
		if err != nil {
			InternalError(err)
			c.String(500, err.Error())
			c.Abort()
			return
		}
		defer rows.Close()
		var results []DistanceResult
		for rows.Next() {
			var result DistanceResult
			rows.Scan(&result.Distance, &result.Date)
			results = append(results, result)
		}
		c.HTML(200, "placeResults", gin.H{"results": results})
	} else {
		c.String(400, "No results found")
	}
}

type DistanceResult struct {
	Distance float64
	Date     time.Time
}
