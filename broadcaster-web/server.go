//go:generate npm run build
//go:generate go-bindata-assetfs dist/...
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/rakyll/statik/fs"

	// import statik filesystem
	_ "github.com/ryex/go-broadcaster/broadcaster-web/client/statik"

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

	CreateAPIRoutes(e, &api)

	logutils.Log.Info("running at localhost:8080")

	e.HideBanner = true
	e.Logger.Fatal(e.Start(":8080"))

}

func CreateAPIRoutes(e *echo.Echo, api *api.Api) {
	g := e.Group("/api")
	g.GET("/library", api.GetLibraryPaths)
	g.GET("/library/:id", api.GetLibraryPath)
	g.PUT("/library", api.PutLibraryPath)
	g.DELETE("/library/:id", api.DeleteLibraryPath)

	g.GET("/track/:id", api.GetTrack)
	g.GET("/tracks", api.GetTracks)
	g.PUT("/track", api.PutTrack)
	g.DELETE("/track/:id", api.DeleteTrack)

	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		return
	}
	ioutil.WriteFile("routes.json", data, 0644)

}

func CreateStaticRoutes(e *echo.Echo, api *api.Api) {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	e.


}
