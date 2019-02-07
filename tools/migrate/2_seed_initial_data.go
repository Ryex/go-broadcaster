package main

import (
	"github.com/go-pg/migrations"
	"github.com/ryex/go-broadcaster/internal/models"
)

func init() {
	roles := []models.Role{
		models.Role{
			ID:    1,
			IDStr: "Admin",
			Perms: map[string]bool{
				"admin": true,
			},
			ParentID: 0,
		},
	}

	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		for _, role := range roles {
			_, err = db.Model(&role).Insert()
			if err != nil {
				return err
			}
		}
		//_, err := db.Exec("")
		return err
	}, func(db migrations.DB) (err error) {
		for _, role := range roles {
			_, err = db.Model(&role).WherePK().Delete()
			if err != nil {
				return err
			}
		}
		//_, err := db.Exec("")
		return err
	})
}
