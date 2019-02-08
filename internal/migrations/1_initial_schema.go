package migrations

import (
	"github.com/go-pg/migrations"
)

func init() {

	upcmd := `
	CREATE TABLE "library_paths" (
	  "id" bigserial,
	  "path" text,
	  "added" timestamptz DEFAULT now(),
	  "last_index" timestamptz,
	  "indexing" boolean,
	  PRIMARY KEY ("id")
	);

	CREATE TABLE "tracks" (
	  "id" bigserial,
	  "title" text,
	  "album" text,
	  "artist" text,
	  "genre" text,
	  "year" bigint,
	  "length" bigint,
	  "bitrate" bigint,
	  "channels" bigint,
	  "samplerate" bigint,
	  "path" text,
	  "added" timestamptz DEFAULT now(),
	  PRIMARY KEY ("id")
	);

	CREATE TABLE "users" (
	  "id" bigserial,
	  "username" text UNIQUE,
	  "password" text,
	  "created_at" timestamptz DEFAULT now(),
	  PRIMARY KEY ("id")
	);

	CREATE TABLE "roles" (
	  "id" bigserial,
	  "id_str" text UNIQUE,
	  "parent_id" bigint,
	  "perms" jsonb,
	  PRIMARY KEY ("id"),
	  FOREIGN KEY ("parent_id") REFERENCES "roles" ("id")
	);

	CREATE TABLE "user_to_roles" (
	  "user_id" bigint,
	  "role_id" bigint,
	  FOREIGN KEY ("user_id") REFERENCES "users" ("id"),
	  FOREIGN KEY ("role_id") REFERENCES "roles" ("id")
	);
	`

	downcmd := `
	DROP TABLE IF EXISTS "library_paths";
	DROP TABLE IF EXISTS "tracks";
	DROP TABLE IF EXISTS "users";
	DROP TABLE IF EXISTS "roles";
	DROP TABLE IF EXISTS "user_to_roles";
	`

	migrations.MustRegisterTx(func(db migrations.DB) error {
		_, err := db.Exec(upcmd)
		return err
	}, func(db migrations.DB) error {
		_, err := db.Exec(downcmd)
		return err
	})
}
