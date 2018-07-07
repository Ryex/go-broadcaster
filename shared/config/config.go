package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DBURL             string   `json:"db_url"`
	DPPort            string   `json:"dp_port"`
	DBDatabase        string   `json:"db_database"`
	DBUser            string   `json:"db_user"`
	DBPassword        string   `json:"db_password"`
	MediaExts         []string `json:"media_exts"`
	Debug             bool     `json:"debug"`
	Development       bool     `json:"development"`
	AuthSecret        string   `json:"auth_secret"`
	AuthTimeoutHours  int      `json:"auth_timeout_hours"`
	UseAssetsFromDisk bool     `json:"use_assets_from_disk"`
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
