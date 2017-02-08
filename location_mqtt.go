package main

import (
	"database/sql"
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lib/pq"
	"log"
	"math"
	"time"
)

type MQTTMsg struct {
	Type                 string             `json:"_type" binding:"required"`
	TrackerId            string             `json:"tid"`
	Accuracy             int                `json:"acc"`
	Battery              int                `json:"batt"`
	Connection           string             `json:"conn"`
	Doze                 ConvertibleBoolean `json:"doze"`
	Latitude             float64            `json:"lat"`
	Longitude            float64            `json:"lon"`
	DeviceTimestampAsInt int64              `json:"tst" binding:"required"`
	DeviceTimestamp      time.Time
	Distance             float64
}

func SubscribeMQTT(quit <-chan bool) error {
	topic := "owntracks/#"
	log.Print("Connecting to MQTT")
	var mqttClientOptions = mqtt.NewClientOptions()
	if configuration.MQTTURL != "" {
		mqttClientOptions.AddBroker(configuration.MQTTURL)
	} else {
		mqttClientOptions.AddBroker("tcp://localhost:1883")
	}
	if configuration.MQTTUsername != "" && configuration.MQTTPassword != "" {
		mqttClientOptions.SetUsername(configuration.MQTTUsername)
		mqttClientOptions.SetPassword(configuration.MQTTPassword)
	}
	mqttClientOptions.SetClientID("growselocator")
	mqttClientOptions.SetAutoReconnect(true)

	mqttClient := mqtt.NewClient(mqttClientOptions)

	mqttClientToken := mqttClient.Connect()
	defer mqttClient.Disconnect(250)
	if mqttClientToken.Wait() && mqttClientToken.Error() != nil {
		log.Printf("Error connecting to mqtt: %v", mqttClientToken.Error())
		return mqttClientToken.Error()
	}
	log.Print("Connected")

	mqttSubscribeToken := mqttClient.Subscribe(topic, 0, handler)
	if mqttSubscribeToken.Wait() && mqttSubscribeToken.Error() != nil {
		log.Printf("Error connecting to mqtt: %v", mqttSubscribeToken.Error())
		mqttClient.Disconnect(250)
		return mqttSubscribeToken.Error()
	}
	select {
	case <-quit:
		log.Print("Unsubscribing")
		mqttUnsubscribeToken := mqttClient.Unsubscribe(topic)
		if mqttUnsubscribeToken.Wait() && mqttUnsubscribeToken.Error() != nil {
			log.Printf("Error unsubscribing from mqtt: %v", mqttUnsubscribeToken.Error())
		}
		log.Print("Closing MQTT")
		return nil
	}
}

func (loc *MQTTMsg) GetRelativeDistance(thisDb *sql.DB) {
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

	cos := (math.Sin(phi1)*math.Sin(phi2)*math.Cos(theta1-theta2) +
		math.Cos(phi1)*math.Cos(phi2))

	cos = math.Max(math.Min(cos, 1.0), -1.0)

	arc := math.Acos(cos)

	// Remember to multiply arc by the radius of the earth
	// in your favorite set of units to get length.
	return arc
}

var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received mqtt message from %v", msg.Topic())
	var locator MQTTMsg
	err := json.Unmarshal([]byte(msg.Payload()), &locator)

	if err != nil {
		log.Printf("Error decoding MQTT message: %v", err)
		log.Print(msg.Payload())
		return
	}
	if locator.Type != "location" {
		log.Printf("Received message is of type %v. Skipping", locator.Type)
		return
	}

	newLocation := false

	locator.DeviceTimestamp = time.Unix(locator.DeviceTimestampAsInt, 0)
	locator.GetRelativeDistance(db)
	dozebool := bool(locator.Doze)
	_, err = db.Exec("insert into locations (timestamp,devicetimestamp,latitude,longitude,accuracy,doze,batterylevel,connectiontype,distance) "+
		"values ($1,$2,$3,$4,$5,$6,$7,$8,$9)",
		time.Now(),
		&locator.DeviceTimestamp,
		&locator.Latitude,
		&locator.Longitude,
		&locator.Accuracy,
		&dozebool,
		&locator.Battery,
		&locator.Connection,
		&locator.Distance)

	switch i := err.(type) {
	case nil:
		newLocation = true
		break
	case *pq.Error:
		log.Printf("Pg error: %v", err)
		log.Printf("Managed to get a duplicate timestamp: %v", locator)
	default:
		log.Printf("%T %v", err, err)
		InternalError(i)
		return
	}
	if newLocation {
		GeocodingWorkQueue <- true
	}

}
