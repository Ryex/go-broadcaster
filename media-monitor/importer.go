package main

import (
	"os"

	"github.com/dhowden/tag"
	"github.com/go-pg/pg"
	"github.com/ryex/go-broadcaster/shared/config"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
)

type Importer struct {
	LibPath models.LibraryPath
	Db      *pg.DB
	Cfg     config.Config
}

func ProcessImport(imp Importer) error {
	logutils.Log.Info("Searching for extensions", imp.Cfg.MediaExts)
	imp.LibPath.SearchWalk(imp.Cfg.MediaExts, getMediaInfo)
	return nil
}

func getMediaInfo(path string) error {
	logutils.Log.Info("Getting metadata for ", path)
	file, err := os.Open(path)
	if err != nil {
		logutils.Log.Error("Could not open file", err)
		return err
	}
	meta, merr := tag.ReadFrom(file)
	if merr != nil {
		logutils.Log.Error("Could not read metadata")
		return merr
	}

	track := new(models.Track)
	track.Title = meta.Title()
	track.Album = meta.Album()
	track.AlbumArtist = meta.AlbumArtist()
	track.Composer = meta.Composer()
	track.Genre = meta.Genre()
	track.Year = meta.Year()

	logutils.Log.Info(track)
	logutils.Log.Info("Raw Metadata: ", meta.Raw())

	return nil
}
