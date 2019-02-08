package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/ryex/go-broadcaster/cmd/gobcast-web/api"
	"github.com/ryex/go-broadcaster/internal/config"
	"github.com/ryex/go-broadcaster/internal/logutils"
	//"github.com/ryex/go-broadcaster/internal/models"
	distfs "github.com/ryex/go-broadcaster/cmd/gobcast-web/client"
)

// "encoding/json"
// "fmt"
// "log"
// "os"
// "strconv"github.com/dgrijalva/jwt-go
// "time"
// "net/http"
//
// "github.com/dhowden/tag"
// "github.com/gorilla/mux"
//"github.com/lib/pq"

const usageText = `Runs a http server for the API of go-broadcaster
Configuration:
	If a file 'config.json' is present in the workign directory it will be loaded
	alternatively the path to 'config.json' can be provided as a command line flag

	additionally the values in 'config.json' can be overridden from the environment
		- GOBROADCASTER_DBURI
		- GOBROADCASTER_DBHOST
		- GOBROADCASTER_DBPORT
		- GOBROADCASTER_DBNAME
		- GOBROADCASTER_DBUSER
		- GOBROADCASTER_DBPASS
		- GOBROADCASTER_DEBUG
		- GOBROADCASTER_AUTHSECRET
		- GOBROADCASTER_AUTHTIMEOUT

	additionally the values in 'config.json' and from the environment
	can be overridden by passing additional command line flags

Usage:
  gobcast-web [args]
Arguments:
`

var cfgFlag string

var dbURIFlag string
var dbHostFlag string
var dbPortFlag int
var dbNameFlag string
var dbUserFlag string
var dbPassFlag string

var authSecretFlag string
var authTimeoutFlag time.Duration

var debugFlag bool

func init() {
	flag.Usage = usage
	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")
	flag.StringVar(&cfgFlag, "config", cfgPath, "Path to the config.json file")
	flag.StringVar(&cfgFlag, "c", cfgPath, "Path to the config.json file")

	flag.StringVar(&dbURIFlag, "dburi", "", "URI to the database")
	flag.StringVar(&dbURIFlag, "U", "", "URI to the database")

	flag.StringVar(&dbHostFlag, "dbhost", "", "Database host name. Can contain :<port>")
	flag.StringVar(&dbHostFlag, "H", "", "Database host name. Can contain :<port>")

	flag.IntVar(&dbPortFlag, "dbport", 0, "Database port")
	flag.IntVar(&dbPortFlag, "P", 0, "Database port")

	flag.StringVar(&dbNameFlag, "dbname", "", "Database name to connect to")
	flag.StringVar(&dbNameFlag, "d", "", "Database name to connect to")

	flag.StringVar(&dbUserFlag, "dbuser", "", "Username to use connecting to the database")

	flag.StringVar(&dbPassFlag, "dbpass", "", "Password to use connecting to the database")

	flag.StringVar(&authSecretFlag, "authsecret", "", "Secret used for signing auth")

	flag.DurationVar(&authTimeoutFlag, "authtimeout", 0,
		"Timeout on login sessions")

	flag.BoolVar(&debugFlag, "debug", false, "enable debug mode")

}

func configParseFlags(cfg *config.Config) *config.Config {

	flag.Parse()

	cfgPath := cfgFlag

	dbURI := dbURIFlag
	dbHost := dbHostFlag
	dbPort := dbPortFlag
	dbName := dbNameFlag
	dbUser := dbUserFlag
	dbPass := dbPassFlag

	authSecret := authSecretFlag
	authTimeout := authTimeoutFlag

	debug := debugFlag

	cfgPath, pathErr := filepath.Abs(cfgPath)
	if pathErr != nil {
		fmt.Println("could not get absolute path for config", pathErr)
	}

	fmt.Println("Loading config from: ", cfgPath)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Println("Error when loading configuration", err)
	}

	cfg = populateConfigEnv(cfg)

	if dbURI != "" {
		cfg.DBURI = dbURI
	}
	if dbHost != "" {
		cfg.DBHost = dbHost
	}
	if dbPort != 0 {
		cfg.DBPort = dbPort
	}
	if dbName != "" {
		cfg.DBName = dbName
	}
	if dbUser != "" {
		cfg.DBUser = dbUser
	}
	if dbPass != "" {
		cfg.DBPassword = dbPass
	}

	if authSecret != "" {
		cfg.AuthSecret = authSecret
	}

	if authTimeout != 0 {
		cfg.AuthTimeout = config.Duration{Duration: authTimeout}
	}

	if !cfg.Debug && debug {
		cfg.Debug = debug
	}
	return cfg
}

func populateConfigEnv(cfg *config.Config) *config.Config {
	pre := "GOBROADCASTER_"
	if dbURI, ok := os.LookupEnv(pre + "DBURI"); ok {
		cfg.DBURI = dbURI
	}
	if dbHost, ok := os.LookupEnv(pre + "DBHOST"); ok {
		cfg.DBHost = dbHost
	}
	if dbPortStr, ok := os.LookupEnv(pre + "DBPORT"); ok {
		if dbPort, err := strconv.Atoi(dbPortStr); err != nil {
			fmt.Println("Error parsing DBPORT from env, bad format: ", dbPortStr)
		} else {
			cfg.DBPort = dbPort
		}
	}
	if dbName, ok := os.LookupEnv(pre + "DBNAME"); ok {
		cfg.DBName = dbName
	}
	if dbUser, ok := os.LookupEnv(pre + "DBUSER"); ok {
		cfg.DBUser = dbUser
	}
	if dbPass, ok := os.LookupEnv(pre + "DBPASS"); ok {
		cfg.DBPassword = dbPass
	}
	if debugStr, ok := os.LookupEnv(pre + "DEBUG"); ok {
		if debug, err := strconv.ParseBool(debugStr); err != nil {
			fmt.Println("Error parsing DEBUG from env, bad format: ", debugStr, err.Error())
		} else {
			cfg.Debug = debug
		}
	}
	if authSecret, ok := os.LookupEnv(pre + "AUTHSECRET"); ok {
		cfg.AuthSecret = authSecret
	}
	if authTimeoutStr, ok := os.LookupEnv(pre + "AUTHTIMEOUT"); ok {
		if authTimeout, err := time.ParseDuration(authTimeoutStr); err != nil {
			fmt.Println("Error parsing AUTHTIMEOUT from env, bad format: ", authTimeoutStr, err.Error())
		} else {
			cfg.AuthTimeout = config.Duration{Duration: authTimeout}
		}
	}
	return cfg
}

func main() {

	cfg := &config.Config{}
	// load config from flags and Env
	cfg = configParseFlags(cfg)

	logutils.SetupLogging("broadcaster-web", cfg.Debug, os.Stdout)
	logutils.Log.Debug(fmt.Sprintf("Using config: %+v", cfg))

	err := cfg.FillEmptyFromURI()
	if err != nil {
		logutils.Log.Errorf("Error loading database settings from URI: %s", err)
	}

	db := pg.Connect(&pg.Options{
		Addr:     cfg.DBHost + ":" + strconv.Itoa(cfg.DBPort),
		Database: cfg.DBName,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
	})
	defer db.Close()

	if cfg.Debug {
		SetupDatabaseQueryLogging(db)
	}

	// TODO get better DB Setup

	a := api.Api{
		DB:          db,
		AuthTimeout: cfg.AuthTimeout.Duration,
		Cfg:         cfg,
	}

	e := echo.New()

	e.Use(middleware.Logger())

	api.RegisterRoutes(e, &a, cfg)

	var fs http.FileSystem = distfs.Dist

	CreateStaticRoutes(e, fs)

	// write routes to file
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		return
	}
	ioutil.WriteFile("routes.json", data, 0644)

	logutils.Log.Info("running at localhost:8080")

	e.HideBanner = true
	e.Logger.Fatal(e.Start(":8080"))

}

func usage() {
	fmt.Print(usageText)
	flag.PrintDefaults()
	os.Exit(2)
}

func errorf(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
}

func exitf(s string, args ...interface{}) {
	errorf(s, args...)
	os.Exit(1)
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(q *pg.QueryEvent) {}

func (d dbLogger) AfterQuery(q *pg.QueryEvent) {
	query, err := q.FormattedQuery()
	if err != nil {
		panic(err)
	}

	logutils.Log.Debugf("%s", query)
}

func SetupDatabaseQueryLogging(db *pg.DB) {
	db.AddQueryHook(dbLogger{})
}
