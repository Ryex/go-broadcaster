package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
)

func (api *Api) GetLibraryPath(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("cant parse id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	q := models.LibraryPathQuery{
		DB: api.DB,
	}

	libp, err := q.GetLibraryPathByID(id)
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

func (api *Api) PutLibraryPath(c echo.Context) error {

	values, ferr := c.FormParams()
	fmt.Println(values.Encode(), ferr)

	q := models.LibraryPathQuery{
		DB: api.DB,
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

func (api *Api) GetLibraryPaths(c echo.Context) error {

	q := models.LibraryPathQuery{
		DB: api.DB,
	}

	paths, count, err := q.GetLibraryPaths(c.QueryParams())

	if err != nil {
		logutils.Log.Error("db query error", err)
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

func (api *Api) DeleteLibraryPath(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("cant parse id", err)
		return c.JSON(http.StatusBadRequest, Responce{
			Err: err,
		})
	}

	q := models.LibraryPathQuery{
		DB: api.DB,
	}

	err = q.DeleteLibraryPathByID(id)

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
