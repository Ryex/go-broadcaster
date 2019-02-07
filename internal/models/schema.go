package models

import (
	"fmt"
	"reflect"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

func init() {
	// Register many to many model so ORM can better recognize m2m relation.
	// This should be done before dependant models are used.
	orm.RegisterTable((*UserToRole)(nil))
}

// Models is a slice listing all modles present for use in Schema operations
var Models = []interface{}{
	(*LibraryPath)(nil),
	(*Track)(nil),
	(*User)(nil),
	(*Role)(nil),
	(*UserToRole)(nil),
}

// CreateSchema creates the database schema useing the go-pg  modles listed
// in the exported Modles slice.
// - db - the databse connection
// - ifne - should the IfNotExists constrant be used in the orm.CreateTableOptions
func CreateSchema(db *pg.DB, ifne bool) error {

	tableOpts := new(orm.CreateTableOptions)
	tableOpts.FKConstraints = true
	tableOpts.IfNotExists = ifne

	for _, model := range Models {
		fmt.Printf("Creating table for model %s\n", reflect.TypeOf(model).Elem().Name())
		err := db.CreateTable(model, tableOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

// DropSchema drops the tables for all models listed in the
// exported Modles slice.
// - db - the database connection
// - ife - should the IfExists constrant be used in the orm.DropTableOptions
// - cascade - should the Cascade option be used in the orm.DropTableOptions
func DropSchema(db *pg.DB, ife bool, cascade bool) error {

	tableOpts := new(orm.DropTableOptions)
	tableOpts.IfExists = ife
	tableOpts.Cascade = cascade

	// itterate in reverse so relation refrences don't casue problems
	for i := len(Models) - 1; i >= 0; i-- {
		model := Models[i]
		fmt.Printf("Droping table for model %s\n", reflect.TypeOf(model).Elem().Name())
		err := db.DropTable(model, tableOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

// ModelNames returns a list of the names of the
// models in the exported Models slice
func ModelNames() []string {
	names := make([]string, len(Models))
	for i, model := range Models {
		t := reflect.TypeOf(model)
		//fmt.Printf("%s, %s\n", t.Elem().Name(), t.Name())
		names[i] = t.Elem().Name()
	}
	return names
}
