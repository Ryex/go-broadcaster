package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	td, _ := time.ParseDuration("24h")
	d := Duration{Duration: td}

	fmt.Printf("Duration Type: %T\n", d)
	fmt.Printf("time.Duration Type: %T\n", td)
	fmt.Printf("Embeded Type access: %T\n", d.Duration)
}

func TestConfigToJSON(t *testing.T) {
	cfg := Config{
		DBInfo: DBInfo{},
	}

	s, err := json.Marshal(cfg)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Json Marsheled cfg:\n %s\n", s)
}

func TestJSONToConfig(t *testing.T) {
	cfgstr := `
	{
	  "db_uri":  "postgres://dbuser:dbpass@localhost:5432/go_broadcaster",
	  "db_host":  "localhost",
		"db_port":  5432,
	  "db_database": "go_broadcaster",
		"db_user":   "dbuser",
		"db_password": "dbpass",
	  "media_exts": [".mp3", ".ogg", ".flac", ".aac"],
	  "debug": false,
	  "development": false,
	  "auth_secret": "OhGodsPleaseChangeMe!",
	  "auth_timeout": "24h"
	}
	`

	cfg := Config{}
	decoder := json.NewDecoder(strings.NewReader(cfgstr))
	err := decoder.Decode(&cfg)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Loaded Config: %+v\n", cfg)

}

func TestParseURI(t *testing.T) {
	cfg := Config{}
	cfg.DBURI = "postgres://dbuser:dbpass@localhost:5432/go_broadcaster"

	if err := cfg.FillEmptyFromURI(); err != nil {
		t.Error(err)
	}

	fmt.Printf("CFG: %+v\n", cfg)
}
