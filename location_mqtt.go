package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"encoding/json"
	"time"
	"github.com/lib/pq"
)

type MQTTMsg struct {
	Type                 string `json:"_type" binding:"required"`
	TrackerId            string `json:"tid"`
	Accuracy             int `json:"acc"`
	Battery              int `json:"batt"`
	Connection           string `json:"conn"`
	Doze                 ConvertibleBoolean `json:"doze"`
	Latitude             float64 `json:"lat"`
	Longitude            float64 `json:"lon"`
	DeviceTimestampAsInt int64 `json:"tst" binding:"required"`
	DeviceTimestamp      time.Time
}

func SubscribeMQTT(quit <-chan bool) error {
	topic := "owntracks/#"
	log.Print("Connecting to MQTT")
	var mqttClientOptions = mqtt.NewClientOptions()
	if (configuration.MQTTURL != "") {
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

var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received mqtt message from %v id %v", msg.Topic(), msg.MessageID())
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

	dozebool := bool(locator.Doze)
	_, err = db.Exec("insert into locations (timestamp,devicetimestamp,latitude,longitude,accuracy,doze,batterylevel,connectiontype) " +
		"values ($1,$2,$3,$4,$5,$6,$7,$8)",
		time.Now(),
		&locator.DeviceTimestamp,
		&locator.Latitude,
		&locator.Longitude,
		&locator.Accuracy,
		&dozebool,
		&locator.Battery,
		&locator.Connection)

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