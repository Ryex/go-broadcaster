package api

import (
	"github.com/go-pg/pg"
)

type Api struct {
	DB *pg.DB
}

type H map[string]interface{}
