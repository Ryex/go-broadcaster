package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"github.com/ryex/go-broadcaster/internal/models"
)

// GET /api/library/id/:id
func (a *Api) GetLibraryPathById(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("cant parse id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	q := models.LibraryPathQuery{
		DB: a.DB,
	}

	libp, err := q.GetLibraryPathById(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"path": libp,
		},
	})

}

// POST /api/library
func (a *Api) PutLibraryPath(c echo.Context) error {

	values, ferr := c.FormParams()
	fmt.Println(values.Encode(), ferr)

	q := models.LibraryPathQuery{
		DB: a.DB,
	}

	libp, err := q.AddLibraryPath(c.FormValue("path"))

	if err != nil {
		return c.JSON(http.StatusInternalServerError, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusCreated, Responce{
		Data: H{
			"created": libp,
		},
	})

}

// GET /api/library
func (a *Api) GetLibraryPaths(c echo.Context) error {

	q := models.LibraryPathQuery{
		DB: a.DB,
	}

	paths, count, err := q.GetLibraryPaths(c.QueryParams())

	if err != nil {
		logutils.Log.Error("db query error %s", err)
		return c.JSON(http.StatusNotFound, Responce{
			Err: err,
		})
	}

	return c.JSON(http.StatusOK, Responce{
		Data: H{
			"paths": paths,
			"count": count,
		},
	})

}

// GELETE /api/library/:id
func (a *Api) DeleteLibraryPath(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("cant parse id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	q := models.LibraryPathQuery{
		DB: a.DB,
	}

	err = q.DeleteLibraryPathById(id)

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
