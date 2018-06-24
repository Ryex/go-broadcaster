package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/go-pg/pg"
	"github.com/ryex/go-broadcaster/shared/config"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
)

func main() {

	logutils.SetupLogging()

	root, _ := os.Getwd()
	cfgPath := filepath.Join(root, "config.json")

	cfgPtr := flag.String("config", cfgPath, "Path to the config.json file")

	flag.Parse()

	cfgPath = *cfgPtr
	// DEBUG
	cfgPath = "..\\config.json"

	cfgPath, pathErr := filepath.Abs(cfgPath)
	if pathErr != nil {
		logutils.Log.Error("could not get absolute path for config")
	}

	logutils.Log.Info("Loading config from: ", cfgPath)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		logutils.Log.Error("error when loading configuration", err)
	}
	logutils.Log.Info("Config Loaded: ", cfg)

	db := pg.Connect(&pg.Options{
		Addr:     cfg.DBURL + ":" + cfg.DPPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
	})
	defer db.Close()

	err = models.CreateSchema(db)
	if err != nil {
		logutils.Log.Error("error loading database schema", err)
	}

	var libPaths []models.LibraryPath
	count, dberr := db.Model(&libPaths).SelectAndCount()
	if dberr != nil {
		logutils.Log.Error("db query error", dberr)
	}

	logutils.Log.Info("Getting Library Paths form database...")
	var importers []Importer
	for i := 0; i < count; i++ {
		imp := new(Importer)
		imp.LibPath = libPaths[i]
		imp.Db = db
		imp.Cfg = cfg
		importers = append(importers, *imp)
		logutils.Log.Info("Building importer for : ", imp.LibPath.Path)
	}

	// Loop indefintely
	//for {
	logutils.Log.Info("Starting import Process")
	for _, im := range importers {
		ProcessImport(im)
	}
	//}

}
