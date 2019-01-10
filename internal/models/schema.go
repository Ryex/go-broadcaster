package models

import "github.com/go-pg/pg"

func CreateSchema(db *pg.DB) error {

	models := []interface{}{
		(*LibraryPath)(nil),
		(*Track)(nil),
		(*User)(nil),
		(*Role)(nil),
	}

	for _, model := range models {
		err := db.CreateTable(model, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
