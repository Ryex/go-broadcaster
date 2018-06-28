package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet"
	"github.com/ryex/go-broadcaster/shared/logutils"
	"github.com/ryex/go-broadcaster/shared/models"
)

func LibraryHandeler(env *Env) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		fmt.Println(req.Method)
		if req.Method == "POST" {
			ferr := req.ParseForm()
			if ferr != nil {
				logutils.Log.Info("could not parse form data")
			}
			for k, v := range req.Form {
				fmt.Println("key:", k)
				fmt.Println("val:", strings.Join(v, ""))
			}
			libPath := new(models.LibraryPath)
			libPath.Path = req.FormValue("libpath")
			ierr := env.DB.Insert(libPath)
			if ierr != nil {
				logutils.Log.Error("Could not add lib path to DB", ierr)
			}
		}

		t, err := env.View.GetTemplate("library.html")
		if err != nil {
			logutils.Log.Error("library template not be loaded", err)
		}

		vars := make(jet.VarMap)

		var libPaths []models.LibraryPath
		count, dberr := env.DB.Model(&libPaths).Limit(20).SelectAndCount()
		if dberr != nil {
			logutils.Log.Error("db query error", dberr)
		}
		vars.Set("libPaths", libPaths)
		vars.Set("libPathsCount", count)

		if err = t.Execute(w, vars, nil); err != nil {
			logutils.Log.Error("error when executing library template", err)
		}

	}
}
