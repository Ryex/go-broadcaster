package migrations

import (
	pgmigrations "github.com/go-pg/migrations"
)

func init() {
	pgmigrations.DefaultCollection = pgmigrations.DefaultCollection.
		DisableSQLAutodiscover(true).
		SetTableName("gobcast_migrations")
}

// Run passes its arguments to github.com/go-pg/migrations#Run
func Run(db pgmigrations.DB, a ...string) (oldVersion, newVersion int64, err error) {
	return pgmigrations.Run(db, a...)
}

// Version retrives and returns the version the database is currently on
func Version(db pgmigrations.DB) (int64, error) {
	return pgmigrations.Version(db)
}

// RegisteredVersions returns a slice of the registered version numbers
func RegisteredVersions() (versions []int64) {
	migrations := pgmigrations.DefaultCollection.Migrations()
	for _, m := range migrations {
		versions = append(versions, m.Version)
	}
	return
}
