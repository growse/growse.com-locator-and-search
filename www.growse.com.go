package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/mailgun/mailgun-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

var (
	db                 *sql.DB
	configuration      Configuration
	gun                mailgun.Mailgun
	oAuthConf          *oauth2.Config
	GeocodingWorkQueue chan bool
)

type Configuration struct {
	DbUser                     string
	DbName                     string
	DbPassword                 string
	DbHost                     string
	DatabaseMigrationsPath     string
	TemplatePath               string
	StaticPath                 string
	GeocodeApiURL              string
	MailgunKey                 string
	Production                 bool
	CookieSeed                 string
	OAuth2CallbackUrl          string
	Domain                     string
	ClientID                   string
	ClientSecret               string
	Port                       int
	MaxDBOpenConnections       int
	MQTTURL                    string
	MQTTUsername               string
	MQTTPassword               string
	SearchIndexRemoteGitUrl    string
	SearchIndexRemoteGitBranch string
	SearchIndexLocalDir        string
}

func InternalError(err error) {
	log.Printf("%v", err)
	debug.PrintStack()
	if configuration.Production {
		m := gun.NewMessage("Sender <blogbot@growse.com>", "ERROR: www.growse.com", fmt.Sprintf("%v\n%v", err, string(debug.Stack())), "sysadmin@growse.com")
		log.Printf("Emailing stack: %v\n", m)
		response, id, _ := gun.Send(m)
		log.Printf("Response ID: %s\n", id)
		log.Printf("Message from server: %s\n", response)
	} else {
		log.Print("Non production, not sending\n")
	}
}

func main() {
	//Flags
	configFile := flag.String("configFile", "config.json", "File path to the JSON configuration")
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	//Config parsing
	file, err := os.Open(*configFile)
	if err != nil {
		log.Fatalf("Unable to open configuration file: %v", err)
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&configuration)

	if err != nil {
		log.Fatalf("Unable to parse configuration file: %v", err)
	}

	if configuration.CookieSeed == "" {
		configuration.CookieSeed = "Wibble"
	}
	if configuration.Port <= 0 {
		configuration.Port = 8000
	}
	oAuthConf = &oauth2.Config{
		ClientID:     configuration.ClientID,
		ClientSecret: configuration.ClientSecret,
		RedirectURL:  configuration.OAuth2CallbackUrl,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}

	//Catch SIGINT & SIGTERM to stop the profiling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	quit := make(chan bool, 1)

	go func() {
		for sig := range c {
			log.Printf("captured %v. Exiting...", sig)
			close(quit)
			close(GeocodingWorkQueue)
			log.Print("Closing manners")
			manners.Close()
		}
		log.Print("Quitting signal listener goroutine.")
	}()

	// Initialize fulltext engine
	BleveInit(configuration.SearchIndexRemoteGitUrl, configuration.SearchIndexRemoteGitBranch, configuration.SearchIndexLocalDir)

	gun = mailgun.NewMailgun("growse.com", configuration.MailgunKey, "")

	// Database time

	connectionString := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable", configuration.DbHost, configuration.DbUser, configuration.DbName)
	if configuration.DbPassword != "" {
		connectionString = fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", configuration.DbHost, configuration.DbUser, configuration.DbName, configuration.DbPassword)
	}

	db, err = sql.Open("postgres", connectionString)

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	} else {
		log.Print("Database connected")
	}

	log.Printf("Setting maximum db connections to %d", configuration.MaxDBOpenConnections)
	db.SetMaxOpenConns(configuration.MaxDBOpenConnections)

	GeocodingWorkQueue = make(chan bool, 100)
	go UpdateLatestLocationWithGeocoding(GeocodingWorkQueue)
	go SubscribeMQTT(quit)

	DoDatabaseMigrations(db, configuration.DatabaseMigrationsPath)
	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatalf("Error closing database: %v", err)
		}
	}()

	//Get the router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	BuildRoutes(router)
	log.Printf("Listening on port %d", configuration.Port)
	err = manners.ListenAndServe(fmt.Sprintf(":%d", configuration.Port), router)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
