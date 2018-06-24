package main

import (
	"github.com/go-pg/pg"
	"github.com/ryex/go-broadcaster/shared/models"
)

type Importer struct {
	Path string
	Db   *pg.DB
}

func (i Importer) ProcessPath(path *models.LibraryPath, db *pg.DB) error {

	return nil
}
