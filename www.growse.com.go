package main

import (
	"database/sql"
	"fmt"
	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime/debug"
	"syscall"
	"time"
)

var (
	db                 *sql.DB
	configuration      Configuration
	oAuthConf          *oauth2.Config
	GeocodingWorkQueue chan bool
)

func InternalError(err error) {
	log.Printf("%v", err)
	if configuration.Production {
		client := sendgrid.NewSendClient(configuration.SendGridKey)
		message := mail.NewSingleEmail(
			mail.NewEmail("Blogface", "blogbot@growse.com"),
			"ERROR: www.growse.com",
			mail.NewEmail("Sysamin", "sysadmin@growse.com"),
			fmt.Sprintf("%v\n%v", err, string(debug.Stack())),
			"")
		log.Printf("Emailing stack: %v\n", message)
		response, err := client.Send(message)
		if err != nil {
			log.Printf("Error sending mail: %v", err)
		} else {
			log.Printf("Email sent: %v\n", response)
		}
	} else {
		log.Println("Non production, not sending")
	}
}

func main() {
	configuration = *getConfiguration()

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
			if quit != nil {
				close(quit)
			}
			if GeocodingWorkQueue != nil {
				close(GeocodingWorkQueue)
			}
			log.Print("Closing manners")
			manners.Close()
		}
		log.Print("Quitting signal listener goroutine.")
	}()

	// Initialize fulltext engine
	pathPattern, err := regexp.Compile(configuration.SearchPathPattern)
	if err != nil {
		log.Printf("Error building regex %v: %v", configuration.SearchPathPattern, err)
	}
	BleveInit(configuration.SearchIndexRoot, pathPattern)

	// Database time
	if configuration.DbHost != "" {
		db, err = setupDatabase(configuration.DbHost, configuration.DbUser, configuration.DbName)
		if err != nil {
			log.Fatalf("Error setting up database")
		}
		GeocodingWorkQueue = make(chan bool, 100)
		go UpdateLatestLocationWithGeocoding(GeocodingWorkQueue)
		if configuration.EnableGeocodingCrawler {
			go GeocodingCrawler(quit)
		}
		go SubscribeMQTT(quit)
		DoDatabaseMigrations(db, configuration.DatabaseMigrationsPath)
	} else {
		log.Print("No database host specified, disabling")
	}
	defer func() {
		if db != nil {
			err := db.Close()
			if err != nil {
				log.Fatalf("Error closing database: %v", err)
			}
		}
	}()

	//Get the router
	if configuration.Production {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.Default()
	BuildRoutes(router)
	log.Printf("Listening on port %d", configuration.Port)
	err = manners.ListenAndServe(fmt.Sprintf(":%d", configuration.Port), router)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func setupDatabase(host string, user string, name string) (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable", host, user, name)
	if configuration.DbPassword != "" {
		connectionString = fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", configuration.DbHost, configuration.DbUser, configuration.DbName, configuration.DbPassword)
	}

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	} else {
		log.Print("Database connected")
	}

	log.Printf("Setting maximum db connections to %d", configuration.MaxDBOpenConnections)
	db.SetMaxOpenConns(configuration.MaxDBOpenConnections)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)
	return db, err
}
