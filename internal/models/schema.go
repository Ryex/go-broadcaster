package models

import "github.com/go-pg/pg"
import "github.com/go-pg/pg/orm"

func CreateSchema(db *pg.DB) error {

	models := []interface{}{
		(*LibraryPath)(nil),
		(*Track)(nil),
		(*User)(nil),
		(*Role)(nil),
	}

	table_opts := new(orm.CreateTableOptions)
	table_opts.FKConstraints = true
	table_opts.IfNotExists = true

	for _, model := range models {
		err := db.CreateTable(model, table_opts)
		if err != nil {
			return err
		}
	}

	return nil
}
