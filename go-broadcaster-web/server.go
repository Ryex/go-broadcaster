package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/ryex/go-broadcaster/go-broadcaster-web/api"
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

	logutils.SetupLogging()

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")

	cfgPtr := flag.String("config", cfgPath, "Path to the config.json file")

	flag.Parse()

	cfgPath = *cfgPtr

	cfgPath, pathErr := filepath.Abs(cfgPath)
	if pathErr != nil {
		logutils.Log.Error("could not get absolute path for config")
	}

	logutils.Log.Info("Loading config from: ", cfgPath)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		logutils.Log.Error("error when loading configuration", err)
	}

	db := pg.Connect(&pg.Options{
		Addr:     cfg.DBURL + ":" + cfg.DPPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
	})
	defer db.Close()

	// TODO get better DB Setup
	models.CreateSchema(db)

	api := api.Api{DB: db}

	e := echo.New()

	e.Use(middleware.Logger())

	CreateAPIRoutes(e, api)

	logutils.Log.Info("running at localhost:8080")

	e.HideBanner = true
	e.Logger.Fatal(e.Start(":8080"))

}

func CreateAPIRoutes(e *echo.Echo, api *api.Api) {
	e.GET("/library", api.GetLibraryPaths)
	e.GET("/library/:id", api.GetLibraryPath)
	e.PUT("/library", api.PutLibraryPath)
	e.DELETE("/library/:id", api.DeleteLibraryPath)

	e.GET("/track/:id", api.GetTrack)
	e.GET("/tracks", api.GetTracks)
	e.PUT("/track", api.PutTrack)
	e.DELETE("/track/:id", api.DeleteTrack)
}
