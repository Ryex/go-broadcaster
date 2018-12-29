//go:generate npm run build
//go:generate go-bindata-assetfs dist/...
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

	"github.com/ryex/go-broadcaster/broadcaster-web/api"
	"github.com/ryex/go-broadcaster/shared/config"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
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

	logutils.SetupLogging("broadcaster-web", cfg.Debug)
	logutils.Log.Info(fmt.Sprintf("Useing config: %+v", cfg))

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
	schemaerr := models.CreateSchema(db)
	if (schemaerr != nil) {
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

	var fs http.FileSystem
	if !cfg.UseAssetsFromDisk {
		logutils.Log.Info("Using binary Asset FS")
		fs = assetFS()
	} else {
		logutils.Log.Info("Using Assets form disk")
		fs = http.Dir("dist")
	}

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

func CreateStaticRoutes(e *echo.Echo, httpfs http.FileSystem) {

	//
	fs := http.FileServer(httpfs)
	//
	// e.GET("/", echo.WrapHandler(fs))
	//
	// e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", fs)))
	//fs := http.FileServer(http.Dir("dist"))
	e.GET("/static/*", func(c echo.Context) error {
		r := c.Request()
		w := c.Response().Writer
		fmt.Println(r.URL.Path)
		fs.ServeHTTP(w, r)
		return nil
	})
	e.GET("/*", func(c echo.Context) error {
		r := c.Request()
		w := c.Response().Writer
		fmt.Println(r.URL.Path)
		r.URL.Path = "/"
		fs.ServeHTTP(w, r)
		return nil
	})

}

func SetupDatabaseQueryLogging(db *pg.DB) {
	db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		query, err := event.FormattedQuery()
		if err != nil {
			panic(err)
		}

		logutils.Log.Debugf("%s %s", time.Since(event.StartTime), query)
	})
}
