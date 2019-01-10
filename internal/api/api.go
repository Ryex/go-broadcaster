package api

import (
	"time"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/ryex/go-broadcaster/internal/config"
)

type Api struct {
	DB          *pg.DB
	AuthTimeout time.Duration
	Cfg         config.Config
}

type H map[string]interface{}

type Responce struct {
	Data H     `json:"data"`
	Err  error `json:"err"`
}

func RegisterRoutes(e *echo.Echo, a *Api, cfg *config.Config) {

	e.POST("/auth", a.Login)

	g := e.Group("/api")

	g.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(cfg.AuthSecret),
	}))

	// Valid Auth Check
	g.GET("/authvalid", a.AuthValid)

	// User
	g.GET("/user", a.GetUsers)
	g.GET("/user/id/:id", a.GetUserById)
	g.GET("/user/name/:name", a.GetUserByName)
	g.POST("/user", a.AddUser)
	g.DELETE("/user/:id", a.DeleteUser)
	g.PUT("/user/id/:id", a.UpdateUserById)
	g.PUT("/user/name/:name", a.UpdateUserByName)

	// Role
	g.GET("/role/id/:id", a.GetRoleById)
	g.GET("/role/name/:name", a.GetRoleByName)
	g.POST("/role", a.AddRole)
	g.DELETE("/role/:id", a.DeleteRoleById)
	g.PUT("/role:/id", a.UpdateRoleById)

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
