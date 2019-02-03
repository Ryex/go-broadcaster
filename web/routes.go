package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

func CreateStaticRoutes(e *echo.Echo, httpfs http.FileSystem) {

	//
	fs := http.FileServer(httpfs)
	//
	// e.GET("/", echo.WrapHandler(fs))
	//
	// e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", fs)))
	//fs := http.FileServer(http.Dir("dist"))
	// e.GET("/static/*", func(c echo.Context) error {
	// 	r := c.Request()
	// 	w := c.Response().Writer
	// 	fmt.Println(r.URL.Path)
	// 	fs.ServeHTTP(w, r)
	// 	return nil
	// })
	e.GET("/*", func(c echo.Context) error {
		r := c.Request()
		w := c.Response().Writer
		fmt.Println(r.URL.Path)
		// r.URL.Path = "/"
		fs.ServeHTTP(w, r)
		return nil
	})

}
