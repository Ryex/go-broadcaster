package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
)

func (api *Api) GetLibraryPath(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("cant parse id", err)
		return err
	}
	libPath := new(models.LibraryPath)
	libPath.Id = id
	dberr := api.DB.Model(libPath).Where("library_path.id = ?", id).Select()
	if dberr != nil {
		logutils.Log.Error("db query error", dberr)
		return dberr
	}

	return c.JSON(http.StatusOK, H{
		"path": libPath,
	})

}

func (api *Api) PutLibraryPath(c echo.Context) error {

	values, ferr := c.FormParams()
	fmt.Println(values.Encode(), ferr)

	libPath := new(models.LibraryPath)
	libPath.Path = c.FormValue("path")
	libPath.Added = time.Now()
	libPath.LastIndex = time.Unix(0, 0)
	if libPath.Path == "" {
		return c.JSON(http.StatusBadRequest, H{
			"message": "missing path",
		})
	}

	ierr := api.DB.Insert(libPath)
	if ierr != nil {
		logutils.Log.Error("Could not add lib path to DB", ierr)
		return c.JSON(http.StatusInternalServerError, H{
			"error": ierr,
		})
	}

	return c.JSON(http.StatusCreated, H{
		"created": libPath,
	})

}

func (api *Api) GetLibraryPaths(c echo.Context) error {

	var libPaths []models.LibraryPath
	count, dberr := api.DB.Model(&libPaths).Limit(20).SelectAndCount()
	if dberr != nil {
		logutils.Log.Error("db query error", dberr)
		return dberr
	}

	return c.JSON(http.StatusOK, H{
		"paths": libPaths,
		"count": count,
	})

}

func (api *Api) DeleteLibraryPath(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logutils.Log.Error("cant parse id", err)
		return err
	}
	libPath := new(models.LibraryPath)
	libPath.Id = id
	_, dberr := api.DB.Model(libPath).Where("library_path.id = ?", id).Delete()
	if dberr != nil {
		logutils.Log.Error("db query error", dberr)
		return dberr
	}

	return c.JSON(http.StatusOK, H{
		"deleted": id,
	})
}
