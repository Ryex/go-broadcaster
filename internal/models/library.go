package models

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/go-pg/pg"
	//"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/urlvalues"
	"github.com/ryex/go-broadcaster/internal/logutils"
	"github.com/ryex/go-broadcaster/internal/utils"
)

type LibraryPath struct {
	Id        int64
	Path      string
	Added     time.Time `sql:"default:now()"`
	LastIndex time.Time
	Indexing  bool
}

func (fp LibraryPath) SearchWalk(extensions []string, cb utils.SearchFunc) error {
	rootPath, aerr := filepath.Abs(fp.Path)
	if aerr != nil {
		fmt.Println("could not get an absolute path")
		return aerr
	}

	werr := utils.WalkSearch(rootPath, extensions, cb)
	if werr != nil {
		fmt.Println("error in search walk")
		return werr
	}
	return nil
}

type LibraryPathQuery struct {
	DB *pg.DB
}

func (lpq *LibraryPathQuery) GetLibraryPaths(queryValues url.Values) (paths []LibraryPath, count int, err error) {
	var pagervalues urlvalues.Values
	err = urlvalues.Decode(queryValues, pagervalues)
	q := lpq.DB.Model(&paths)
	count, err = q.Apply(urlvalues.Pagination(pagervalues)).SelectAndCount()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (lpq *LibraryPathQuery) GetLibraryPathById(id int64) (lp *LibraryPath, err error) {
	lp = new(LibraryPath)
	err = lpq.DB.Model(lp).Where("library_path.id = ?", id).Select()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (lpq *LibraryPathQuery) AddLibraryPath(path string) (lp *LibraryPath, err error) {
	if path == "" {
		err = errors.New("empty path")
		return
	}
	lp = new(LibraryPath)
	lp.Path = path
	lp.Added = time.Now()
	lp.LastIndex = time.Unix(0, 0)
	err = lpq.DB.Insert(lp)
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}

func (lpq *LibraryPathQuery) DeleteLibraryPathById(id int64) (err error) {
	lp := new(LibraryPath)
	_, err = lpq.DB.Model(lp).Where("library_path.id = ?", id).Delete()
	if err != nil {
		logutils.Log.Error("db query error", err)
	}
	return
}
