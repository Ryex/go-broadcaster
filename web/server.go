package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/ryex/go-broadcaster/internal/api"
	"github.com/ryex/go-broadcaster/internal/config"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"github.com/ryex/go-broadcaster/internal/models"
	distfs "github.com/ryex/go-broadcaster/web/filesystem"
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

const usageText = `This runs a http server for the API of go-broadcaster
Usage:
  webapi [-config path/to/config.json]
`

func main() {

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")

	cfgPtr := flag.String("config", cfgPath, "Path to the config.json file")

	flag.Parse()

	cfgPath = *cfgPtr

	cfgPath, pathErr := filepath.Abs(cfgPath)
	if pathErr != nil {
		fmt.Println("could not get absolute path for config", pathErr)
	}

	fmt.Println("Loading config from: ", cfgPath)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Println("Error when loading configuration", err)
	}

	logutils.SetupLogging("broadcaster-web", cfg.Debug, os.Stdout)
	logutils.Log.Info(fmt.Sprintf("Using config: %+v", cfg))

	db := pg.Connect(&pg.Options{
		Addr:     cfg.DBURL + ":" + cfg.DPPort,
		Database: cfg.DBDatabase,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
	})
	defer db.Close()

	if cfg.Debug {
		SetupDatabaseQueryLogging(db)
	}

	// TODO get better DB Setup
	logutils.Log.Info("Setting up database Schema")
	schemaerr := models.CreateSchema(db, true)
	if schemaerr != nil {
		logutils.Log.Error("Error setting up database Schema", schemaerr)
	}

	a := api.Api{
		DB:          db,
		AuthTimeout: time.Hour * time.Duration(cfg.AuthTimeoutHours),
		Cfg:         cfg,
	}

	e := echo.New()

	e.Use(middleware.Logger())

	api.RegisterRoutes(e, &a, &cfg)

	var fs http.FileSystem = distfs.Dist

	CreateStaticRoutes(e, fs)

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
