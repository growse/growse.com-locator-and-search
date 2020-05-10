package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLocationConvertsToOTLoc(t *testing.T) {
	location := Location{
		Latitude:        15,
		Longitude:       20,
		Timestamp:       time.Date(2010, 2, 3, 4, 5, 6, 0, time.UTC),
		DeviceTimestamp: time.Date(2010, 2, 3, 4, 5, 6, 0, time.UTC),
		Accuracy:        5,
		Distance:        20,
		DeviceID:        "",
		Geocoding:       "",
	}
	assert.Equal(t, location.toOT(), OTPos{
		Tst:  1265169906,
		Acc:  5,
		Type: "location",
		Alt:  0,
		Lon:  20,
		Vac:  0,
		Vel:  0,
		Lat:  15,
		Addr: "",
	})
}
