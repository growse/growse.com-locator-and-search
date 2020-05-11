package main

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/spf13/viper"
	"log"
)

type Configuration struct {
	DbUser                 string
	DbName                 string
	DbPassword             string
	DbHost                 string
	DatabaseMigrationsPath string
	GeocodeApiURL          string
	MailgunKey             string
	Production             bool
	CookieSeed             string
	OAuth2CallbackUrl      string
	Domain                 string
	ClientID               string
	ClientSecret           string
	Port                   int
	MaxDBOpenConnections   int
	MQTTURL                string
	MQTTUsername           string
	MQTTPassword           string
	SearchIndexRoot        string
	SearchPathPattern      string
	AllowedAuthUsers       string
}

func getConfiguration() *Configuration {
	viper.SetConfigName("www-growse-com.conf")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/")
	//Config parsing
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Unable to open configuration file: %v", err)
	}

	defaultConfig := &Configuration{
		DbUser:                 "",
		DbName:                 "",
		DbPassword:             "",
		DbHost:                 "",
		DatabaseMigrationsPath: "databasemigrations",
		GeocodeApiURL:          "",
		MailgunKey:             "",
		Production:             false,
		CookieSeed:             getRandomString(),
		OAuth2CallbackUrl:      "https://www.growse.com/oauth2callback",
		Domain:                 "growse.com",
		ClientID:               "",
		ClientSecret:           "",
		Port:                   8080,
		MaxDBOpenConnections:   0,
		MQTTURL:                "",
		MQTTUsername:           "",
		MQTTPassword:           "",
		SearchIndexRoot:        "/var/www/growse-jekyll",
		SearchPathPattern:      "\\d{4}/\\d{2}/\\d{2}/.+?\\.html$",
		AllowedAuthUsers:       "growse@gmail.com",
	}
	err = viper.Unmarshal(&defaultConfig)
	if err != nil {
		log.Fatalf("Unable to parse configuration file: %v", err)
	}
	return defaultConfig
}

func getRandomString() string {
	c := 20
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("Error generating random data: %v", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
