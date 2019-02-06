package models

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-pg/pg"
	//"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/urlvalues"
	"github.com/ryex/go-broadcaster/internal/logutils"
	taglib "github.com/wtolson/go-taglib"
)

type Track struct {
	Id         int64
	Title      string
	Album      string
	Artist     string
	Genre      string
	Year       int
	Length     time.Duration
	Bitrate    int
	Channels   int
	Samplerate int
	Path       string
	Added      time.Time `sql:"default:now()"`
}

func NewTrack(path string) (t *Track, err error) {
	t = new(Track)
	file, err := taglib.Read(path)
	if err != nil {
		logutils.Log.Error("Could not open file", err)
		return
	}
	defer file.Close()

	t.Path = path
	t.Title = file.Title()
	t.Album = file.Album()
	t.Artist = file.Artist()
	t.Genre = file.Genre()
	t.Year = file.Year()
	t.Bitrate = file.Bitrate()
	t.Channels = file.Channels()
	t.Length = file.Length()
	t.Samplerate = file.Samplerate()
	t.Added = time.Now()
	return
}

func (t Track) String() string {
	return fmt.Sprintf("{ Title: %v, Album: %v, Genre: %v, Year: %v, Length: %v, Bitrate: %v, Channels: %v, Samplerate: %v, Path: %v}",
		t.Title, t.Album, t.Genre, t.Year, t.Length, t.Bitrate, t.Channels, t.Samplerate, t.Path)
}

type TrackQuery struct {
	DB *pg.DB
}

func (tq *TrackQuery) GetTrackById(id int64) (t *Track, err error) {
	t = new(Track)
	err = tq.DB.Model(t).Where("track.id = ?", id).Select()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

func (tq *TrackQuery) GetTracks(queryValues url.Values) (tracks []Track, count int, err error) {
	var pagervalues urlvalues.Values
	err = urlvalues.Decode(queryValues, pagervalues)
	q := tq.DB.Model(&tracks)
	count, err = q.Apply(urlvalues.Pagination(pagervalues)).SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

func (tq *TrackQuery) AddTrack(path string) (t *Track, err error) {
	if path == "" {
		err = errors.New("empty path")
		return
	}
	t, err = NewTrack(path)
	err = tq.DB.Insert(t)
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}

// DeleteTrackById removes a track from the database useing the Id
func (tq *TrackQuery) DeleteTrackById(id int64) (err error) {
	t := new(Track)
	_, err = tq.DB.Model(t).Where("track.id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error %s", err)
	}
	return
}
