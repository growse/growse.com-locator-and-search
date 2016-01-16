package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"time"
)

type KalmanLocation struct {
	Id              int
	Latitude        float64
	Longitude       float64
	Timestamp       time.Time
	Accuracy        float64
	Distance        float64
	KalmanLatitude  float64
	KalmanLongitude float64
	KalmanAccuracy  float64
	KalmanDistance  float64
}

func KalmanHandler(c *gin.Context) {
	c.String(200, "Yay")
}

func RecalculateKalmanHandler(c *gin.Context) {
	DoKalmanFiltering(db)
	c.Redirect(302, "/auth/where/kalman")
}

/*
Commit the kalman data for the given location back to the database
*/
func (kalmanLocation KalmanLocation) Persist(tx *sql.Tx) {
	log.Printf("Persisting: %v", kalmanLocation)
	_, err := tx.Exec("update locations set kalmanlatitude=$1, kalmanlongitude=$2, kalmanaccuracy=$3, kalmandistance=$4 where id=$5", kalmanLocation.KalmanLatitude, kalmanLocation.Longitude, kalmanLocation.KalmanAccuracy, kalmanLocation.KalmanDistance, kalmanLocation.Id)
	if err != nil {
		log.Fatalf("Fatal persisting location: %v: %v", kalmanLocation, err)
	}
}

/*
Kind of taken from http://stackoverflow.com/questions/1134579/smooth-gps-data
*/
func DoKalmanFiltering(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Fatal starting db transaction: %v", err)
	}
	rows, err := tx.Query("select id,latitude,longitude,devicetimestamp,accuracy,distance from locations order by devicetimestamp asc")
	var locations []KalmanLocation
	for rows.Next() {
		kalmanLocation := KalmanLocation{}
		rows.Scan(
			&kalmanLocation.Id,
			&kalmanLocation.Latitude,
			&kalmanLocation.Longitude,
			&kalmanLocation.Timestamp,
			&kalmanLocation.Accuracy,
			&kalmanLocation.Distance)
		locations = append(locations, kalmanLocation)
	}
	var latState float64
	var longState float64
	var timestampState time.Time
	var variance float64
	var minAccuracy float64
	var Q float64
	variance = -1
	minAccuracy = 1
	Q = 3

	for _, location := range locations {
		prevLat := latState
		prevLong := longState
		if location.Accuracy < minAccuracy {
			minAccuracy = location.Accuracy
		}
		if variance < 0 {
			log.Print("Initial point")
			latState = location.Latitude
			longState = location.Longitude
			timestampState = location.Timestamp
			variance = location.Accuracy * location.Accuracy
		} else {

			timeDelta := location.Timestamp.Sub(timestampState)
			log.Printf("New point, timedelta: %v", timeDelta)
			if timeDelta > 0 {
				variance += timeDelta.Seconds() * Q * (Q / 1000)
				log.Printf("Adjusted variance to %v", variance)
				timestampState = location.Timestamp
			}
			var K float64
			K = variance / (variance + (location.Accuracy * location.Accuracy))
			latState += K * (location.Latitude - latState)
			longState += K * (location.Longitude - longState)
			variance = (1 - K) * variance
			log.Printf("Values moved from %v,%v to %v,%v K=%v new variance=%v", location.Latitude, location.Longitude, latState, longState, K, variance)
		}
		if prevLat != 0 && prevLong != 0 {
			location.KalmanDistance = 6378100 * DistanceOnUnitSphere(latState, longState, prevLat, prevLong)
			log.Printf("Distance on a sphere between %v,%v and %v,%v was %v", latState, longState, prevLat, prevLong, location.KalmanDistance)
		}
		location.KalmanAccuracy = math.Sqrt(variance)
		location.KalmanLatitude = latState
		location.KalmanLongitude = longState
		location.Persist(tx)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Fatal committing transaction: %v", err)
	}
}
