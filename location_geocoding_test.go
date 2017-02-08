package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeocodingShouldDecodeLocalityToTheRightName(t *testing.T) {
	testlocation := "{\"status\": \"OK\", \"results\": [{\"types\": [\"locality\", \"political\"], \"geometry\": {\"bounds\": {\"northeast\": {\"lat\": 51.7829531, \"lng\": -0.4110989}, \"southwest\": {\"lat\": 51.7245271, \"lng\": -0.517344}}, \"location\": {\"lat\": 51.753241, \"lng\": -0.448632}, \"viewport\": {\"northeast\": {\"lat\": 51.7829531, \"lng\": -0.4110989}, \"southwest\": {\"lat\": 51.7245271, \"lng\": -0.517344}}, \"location_type\": \"APPROXIMATE\"}, \"place_id\": \"ChIJwRrXXNBAdkgRsvLfzx9_4PA\", \"formatted_address\": \"Hemel Hempstead, UK\", \"address_components\": [{\"types\": [\"locality\", \"political\"], \"long_name\": \"Hemel Hempstead\", \"short_name\": \"Hemel Hempstead\"}, {\"types\": [\"administrative_area_level_3\", \"political\"], \"long_name\": \"Dacorum District\", \"short_name\": \"Dacorum District\"}, {\"types\": [\"administrative_area_level_2\", \"political\"], \"long_name\": \"Hertfordshire\", \"short_name\": \"Hertfordshire\"}, {\"types\": [\"administrative_area_level_1\", \"political\"], \"long_name\": \"England\", \"short_name\": \"England\"}, {\"types\": [\"country\", \"political\"], \"long_name\": \"United Kingdom\", \"short_name\": \"GB\"}]}]}"
	location := Location{Geocoding: testlocation}
	name := location.Name()
	assert.Equal(t, "Hemel Hempstead", name)
}

func TestGeocodingShouldDecodePostalTownToTheRightName(t *testing.T) {
	testlocation := "{\"status\": \"OK\", \"results\": [{\"types\": [\"locality\", \"political\"], \"geometry\": {\"bounds\": {\"northeast\": {\"lat\": 51.7829531, \"lng\": -0.4110989}, \"southwest\": {\"lat\": 51.7245271, \"lng\": -0.517344}}, \"location\": {\"lat\": 51.753241, \"lng\": -0.448632}, \"viewport\": {\"northeast\": {\"lat\": 51.7829531, \"lng\": -0.4110989}, \"southwest\": {\"lat\": 51.7245271, \"lng\": -0.517344}}, \"location_type\": \"APPROXIMATE\"}, \"place_id\": \"ChIJwRrXXNBAdkgRsvLfzx9_4PA\", \"formatted_address\": \"Hemel Hempstead, UK\", \"address_components\": [{\"types\": [\"locality\", \"political\"], \"long_name\": \"Hemel Hempstead\", \"short_name\": \"Hemel Hempstead\"}, {\"types\": [\"postal_town\"], \"long_name\": \"Hemelty Hempstead\", \"short_name\": \"Hemel Hempstead\"}, {\"types\": [\"administrative_area_level_3\", \"political\"], \"long_name\": \"Dacorum District\", \"short_name\": \"Dacorum District\"}, {\"types\": [\"administrative_area_level_2\", \"political\"], \"long_name\": \"Hertfordshire\", \"short_name\": \"Hertfordshire\"}, {\"types\": [\"administrative_area_level_1\", \"political\"], \"long_name\": \"England\", \"short_name\": \"England\"}, {\"types\": [\"country\", \"political\"], \"long_name\": \"United Kingdom\", \"short_name\": \"GB\"}]}]}"
	location := Location{Geocoding: testlocation}
	name := location.Name()
	assert.Equal(t, "Hemelty Hempstead", name)
}

func TestGeocodingShouldDecodeAdministrativeAreaToTheRightName(t *testing.T) {
	testlocation := "{\"status\": \"NOT OK\", \"results\": [{\"types\": [\"locality\", \"political\"], \"geometry\": {\"bounds\": {\"northeast\": {\"lat\": 51.7829531, \"lng\": -0.4110989}, \"southwest\": {\"lat\": 51.7245271, \"lng\": -0.517344}}, \"location\": {\"lat\": 51.753241, \"lng\": -0.448632}, \"viewport\": {\"northeast\": {\"lat\": 51.7829531, \"lng\": -0.4110989}, \"southwest\": {\"lat\": 51.7245271, \"lng\": -0.517344}}, \"location_type\": \"APPROXIMATE\"}, \"place_id\": \"ChIJwRrXXNBAdkgRsvLfzx9_4PA\", \"formatted_address\": \"Hemel Hempstead, UK\", \"address_components\": [{\"types\": [\"locality\", \"political\"], \"long_name\": \"Hemel Hempstead\", \"short_name\": \"Hemel Hempstead\"}, {\"types\": [\"postal_town\"], \"long_name\": \"Hemel Hempstead\", \"short_name\": \"Hemel Hempstead\"}, {\"types\": [\"administrative_area_level_3\", \"political\"], \"long_name\": \"Dacorum District\", \"short_name\": \"Dacorum District\"}, {\"types\": [\"administrative_area_level_2\", \"political\"], \"long_name\": \"Hertfordshire\", \"short_name\": \"Hertfordshire\"}, {\"types\": [\"administrative_area_level_1\", \"political\"], \"long_name\": \"England\", \"short_name\": \"England\"}, {\"types\": [\"country\", \"political\"], \"long_name\": \"United Kingdom\", \"short_name\": \"GB\"}]}]}"
	location := Location{Geocoding: testlocation}
	name := location.Name()
	assert.Equal(t, "Unknown", name)
}
