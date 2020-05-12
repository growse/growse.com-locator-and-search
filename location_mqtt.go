package main

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lib/pq"
	"log"
	"time"
)

type MQTTMsg struct {
	Type                 string             `json:"_type" binding:"required"`
	TrackerId            string             `json:"tid"`
	Accuracy             float32            `json:"acc"`
	VerticalAccuracy     float32            `json:"vac"`
	Battery              int                `json:"batt"`
	Connection           string             `json:"conn"`
	Doze                 ConvertibleBoolean `json:"doze"`
	Latitude             float64            `json:"lat"`
	Longitude            float64            `json:"lon"`
	Speed                float32            `json:"vel"`
	Altitude             float32            `json:"alt"`
	DeviceTimestampAsInt int64              `json:"tst" binding:"required"`
	DeviceTimestamp      time.Time
}

var topic = "owntracks/growse/nexus5"

func SubscribeMQTT(quit <-chan bool) error {
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
	mqttClientOptions.SetConnectionLostHandler(connectionLostHandler)
	mqttClientOptions.SetOnConnectHandler(onConnectHandler)
	mqttClient := mqtt.NewClient(mqttClientOptions)

	mqttClientToken := mqttClient.Connect()
	defer mqttClient.Disconnect(250)

	if mqttClientToken.Wait() && mqttClientToken.Error() != nil {
		log.Printf("Error connecting to mqtt: %v", mqttClientToken.Error())
		return mqttClientToken.Error()
	}
	log.Print("MQTT Connected")

	err := subscribeToMQTT(mqttClient, topic, handler)
	if err != nil {
		return err
	}

	select {
	case <-quit:
		log.Print("MQTT Unsubscribing")
		mqttUnsubscribeToken := mqttClient.Unsubscribe(topic)
		if mqttUnsubscribeToken.Wait() && mqttUnsubscribeToken.Error() != nil {
			log.Printf("Error unsubscribing from mqtt: %v", mqttUnsubscribeToken.Error())
		}
		log.Print("Closing MQTT")
		return nil
	}
}
func subscribeToMQTT(mqttClient mqtt.Client, topic string, handler mqtt.MessageHandler) error {
	log.Printf("MQTT Subscribing to %v", topic)
	mqttSubscribeToken := mqttClient.Subscribe(topic, 0, handler)
	if mqttSubscribeToken.Wait() && mqttSubscribeToken.Error() != nil {
		log.Printf("Error connecting to mqtt: %v", mqttSubscribeToken.Error())
		mqttClient.Disconnect(250)
		return mqttSubscribeToken.Error()
	}
	log.Printf("MQTT Subscribed to %v", topic)
	return nil
}

var onConnectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Print("MQTT Connected!")
	subscribeToMQTT(client, topic, handler)
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("MQTT Connection lost: %v", err)
}

var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	defer timeTrack(time.Now())
	log.Printf("Received mqtt message from %v", msg.Topic())
	var locator MQTTMsg
	err := json.Unmarshal(msg.Payload(), &locator)

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
	_, err = db.Exec(
		"insert into locations "+
			"(timestamp,devicetimestamp,accuracy,doze,batterylevel,connectiontype,point, altitude, verticalaccuracy, speed) "+
			"values ($1,$2,$3,$4,$5,$6, ST_SetSRID(ST_MakePoint($7, $8), 4326), $9, $10, $11)",

		time.Now(),
		locator.DeviceTimestamp,
		locator.Accuracy,
		dozebool,
		locator.Battery,
		locator.Connection,
		locator.Longitude,
		locator.Latitude,
		locator.Altitude,
		locator.VerticalAccuracy,
		locator.Speed,
	)

	switch i := err.(type) {
	case nil:
		newLocation = true
		break
	case *pq.Error:
		log.Printf("Pg error: %v", err)
		log.Printf("Locator struct: %v", locator)
	default:
		log.Printf("%T %v", err, err)
		InternalError(i)
		return
	}
	if newLocation {
		GeocodingWorkQueue <- true
	}
}
