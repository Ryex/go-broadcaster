package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/CloudyKit/jet"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/ryex/go-broadcaster/shared/config"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
)

// "encoding/json"
// "fmt"
// "log"
// "os"
// "strconv"
// "time"
// "net/http"
//
// "github.com/dhowden/tag"
// "github.com/gorilla/mux"
//"github.com/lib/pq"

type Env struct {
	DB   *pg.DB
	View *jet.Set
}

func main() {

	logutils.SetupLogging()

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")
	view := jet.NewHTMLSet(filepath.Join(root, "views"))

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

	env := Env{DB: db, View: view}

	// TODO  remove after development
	env.View.SetDevelopmentMode(true)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/library", LibraryHandeler(&env))

	logutils.Log.Info("running at localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}

}
