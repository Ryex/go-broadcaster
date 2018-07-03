package main

import (
	"github.com/go-pg/pg"
	"github.com/ryex/go-broadcaster/shared/config"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
	taglib "github.com/wtolson/go-taglib"
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
	file, err := taglib.Read(path)
	if err != nil {
		logutils.Log.Error("Could not open file", err)
		return err
	}
	defer file.Close()

	track := new(models.Track)
	track.Path = path
	track.Title = file.Title()
	track.Album = file.Album()
	track.Artist = file.Artist()
	track.Genre = file.Genre()
	track.Year = file.Year()
	track.Bitrate = file.Bitrate()
	track.Channels = file.Channels()
	track.Length = file.Length()
	track.Samplerate = file.Samplerate()

	//file.Track()

	logutils.Log.Info(track)
	//logutils.Log.Info("Raw Metadata: ", meta.Raw())
	logutils.Log.Info("Track Comment: ", file.Comment())

	return nil
}
