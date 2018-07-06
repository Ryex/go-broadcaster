package api

import (
	"github.com/go-pg/pg"
	"github.com/labstack/echo"
)

type Api struct {
	DB *pg.DB
}

type H map[string]interface{}

type Responce struct {
	Data H     `json:"data"`
	Err  error `json:"err"`
}

func RegisterRoutes(e *echo.Echo, a *Api) {

	g := e.Group("/api")

	// User
	g.GET("/user/id/:id", a.GetUserById)
	g.GET("/user/name/:name", a.GetUserByName)
	g.POST("/user", a.AddUser)
	g.DELETE("/user/:id", a.DeleteUser)

	// Role
	g.GET("/role/id/:id", a.GetRoleById)
	g.GET("/role/name/:name", a.GetRoleByName)
	g.POST("/role", a.AddRole)
	g.DELETE("/role/:id", a.DeleteRole)
	g.PUT("/role:/id", a.UpdateRole)

	// Library Path
	g.GET("/library", a.GetLibraryPaths)
	g.GET("/library/id/:id", a.GetLibraryPathById)
	g.POST("/library", a.PutLibraryPath)
	g.DELETE("/library/:id", a.DeleteLibraryPath)

	// Track
	g.GET("/track/id/:id", a.GetTrackById)
	g.GET("/track", a.GetTracks)
	g.POST("/track", a.AddTrack)
	g.DELETE("/track/:id", a.DeleteTrack)

}
