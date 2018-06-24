package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DBURL      string
	DPPort     string
	DBUser     string
	DBPassword string
}

func LoadConfig(filename string) (Config, error) {

	cfg := Config{}
	file, err := os.Open(filename)
	if err != nil {
		return cfg, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	return cfg, err

}
