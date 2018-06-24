package main

import (
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

	err = models.CreateSchema(db)
	if err != nil {
		logutils.Log.Error("error loading database schema", err)
	}

}
