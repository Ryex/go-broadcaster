package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"github.com/ryex/go-broadcaster/internal/models"
)

// GET /api/track/id/:id
func (a *Api) GetTrackById(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return err
	}

	q := models.TrackQuery{
		DB: a.DB,
	}
	t, err := q.GetTrackById(id)

	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"track": t,
		},
	})
}

// GET /api/tracks
func (a *Api) GetTracks(c echo.Context) error {
	q := models.TrackQuery{
		DB: a.DB,
	}

	tracks, count, err := q.GetTracks(c.QueryParams())
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"tracks": tracks,
			"count":  count,
		},
	})
}

// Mostly for debug purposes not really intended for use
// POST /api/track
func (a *Api) AddTrack(c echo.Context) error {

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

	err = a.DB.Insert(track)
	if err != nil {
		logutils.Log.Error("Could not add track to DB", err)
		return err
	}

	return c.JSON(http.StatusCreated, Responce{
		Data: H{
			"created": track,
		},
	})

}

// DELETE /api/track/:id
func (a *Api) DeleteTrack(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("Error parsing id", err)
		return err
	}

	q := models.TrackQuery{
		DB: a.DB,
	}

	err = q.DeleteTrackById(id)

	if err != nil {
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"deleted": id,
		},
	})
}
