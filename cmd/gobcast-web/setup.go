package main

import (
	"time"

	"github.com/labstack/echo"
	//"github.com/ryex/go-broadcaster/internal/models"
)

// SetupContext hold information about the server setup
type SetupContext struct {
	LastChecked time.Time
}

// Valid checks if the setup of the server and it's database is valid
func (sc *SetupContext) Valid() bool {
	return true
}

type GobcastContext struct {
	echo.Context
	GBSetup SetupContext
}
