package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
)

func (api *Api) GetTrack(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return err
	}
	track := new(models.Track)
	track.Id = id
	dberr := api.DB.Model(track).Where("track.id = ?", id).Select()
	if dberr != nil {
		logutils.Log.Error("db query error", dberr, count)
		return dberr
	}

	return c.JSON(http.StatusOK, H{
		"track": track,
	})
}

func (api *Api) GetTracks(c echo.Context) error {
	var tracks []models.Track

	q := api.DB.Model(&tracks)
	q = q.Apply(orm.Pagination(c.QueryParams()))

	count, dberr := q.SelectAndCount()

	if dberr != nil {
		logutils.Log.Error("db query error", dberr)
		return dberr
	}

	return c.JSON(http.StatusOK, H{
		"tracks": tracks,
		"count":  count,
	})
}

func (api *Api) PutTrack(c echo.Context) error {

	track := new(models.Track)
	track.Path = c.FormValue("path")
	track.Title = c.FormValue("title")
	track.Album = c.FormValue("album")
	track.Artist = c.FormValue("artist")
	track.Genre = c.FormValue("genre")

	year, err := strconv.Atoi(c.FormValue("year"))
	if err != nil {
		logutils.Log.Error("Error parsing year", err)
		return c.JSON(http.StatusBadRequest, H{
			"error": err,
		})
	}
	track.Year = year

	bitrate, err := strconv.Atoi(c.FormValue("bitrate"))
	if err != nil {
		logutils.Log.Error("Error parsing bitrate", err)
		return err
	}
	track.Bitrate = bitrate

	channels, err := strconv.Atoi(c.FormValue("channels"))
	if err != nil {
		logutils.Log.Error("Error parsing channels", err)
		return err
	}
	track.Channels = channels

	length, err := time.ParseDuration(c.FormValue("length"))
	if err != nil {
		logutils.Log.Error("Error parsing length", err)
		return err
	}
	track.Length = length

	samplerate, err := strconv.Atoi(c.FormValue("samplerate"))
	if err != nil {
		logutils.Log.Error("Error parsing samplerate", err)
		return err
	}
	track.Samplerate = samplerate

	track.Added = time.Now()

	err = api.DB.Insert(track)
	if err != nil {
		logutils.Log.Error("Could not add track to DB", err)
		return err
	}

	return c.JSON(http.StatusCreated, H{
		"created": track,
	})

}

func (api *Api) DeleteTrack(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return err
	}

	track := new(models.Track)
	track.Id = id
	_, dberr := api.DB.Model(track).Where("track.id = ?", id).Delete()
	if dberr != nil {
		logutils.Log.Error("db query error", dberr)
		return dberr
	}

	return c.JSON(http.StatusOK, H{
		"deleted": id,
	})
}
