package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DBURL      string   `json:"DBURL"`
	DPPort     string   `json:"DPPort"`
	DBUser     string   `json:"DBUser"`
	DBPassword string   `json:"DBPassword"`
	MediaExts  []string `json:"MediaExts"`
	Debug      bool     `json:Debug`
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
