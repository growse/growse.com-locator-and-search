package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"log"
)

type MQTTMsg struct {
	Type                 string `json: "_type" binding: "required"`
	TrackerId            string `json: "tid" binding: "required"`
	Accuracy             int `json: "acc" binding: "required"`
	Battery              int `json: "batt" binding: "required"`
	Connection           string `json: "conn" binding: "required"`
	Doze                 bool `json: "doze" binding: "required"`
	Latitude             float64 `json: "lat" binding: "required"`
	Longitude            float64 `json: "lon" binding: "required"`
	DeviceTimestampAsInt int `json: "tst" binding:"required"`
}

func SubscribeMQTT() error {

	log.Print("Connecting to MQTT")
	var mqttClientOptions = mqtt.NewClientOptions()
	mqttClientOptions.AddBroker("tcp://localhost:1833")
	mqttClientOptions.SetClientID("growselocator")
	mqttClientOptions.SetAutoReconnect(true)

	var mqttClient = mqtt.NewClient(mqttClientOptions)

	var mqttClientToken = mqttClient.Connect()
	if mqttClientToken.Wait() && mqttClientToken.Error() != nil {

		log.Printf("Error connecting to mqtt: %v", mqttClientToken.Error())
		mqttClient.Disconnect(250)
		return mqttClientToken.Error()
	}
	log.Print("Connected")
	defer mqttClient.Disconnect(250)
	mqttClient.Subscribe("owntracks/#", 0, handler)
	return nil
}

var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("TOPIC: %s\n", msg.Topic())
	log.Printf("MSG: %s\n", msg.Payload())
}