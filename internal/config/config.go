package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type DBInfo struct {
	DBURI      string `json:"db_uri"`
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
}

func parseDBURI(uriStr string) (*DBInfo, error) {
	dbi := &DBInfo{}
	dbi.DBURI = uriStr
	if uriStr != "" {
		uri, err := url.Parse(uriStr)
		if err != nil {
			return dbi, err
		}
		//fmt.Printf("%+v\n", uri)

		if uri.Scheme != "postgres" {
			err = fmt.Errorf("Only the 'postgres' scheme is supported")
			return dbi, err
		}

		hp := uri.Host
		if strings.Contains(hp, ":") {
			pair := strings.Split(hp, ":")
			if len(pair) != 2 {
				err = fmt.Errorf("Malformed host '%s'", hp)
				return dbi, err
			}
			dbi.DBHost = pair[0]
			port, err := strconv.Atoi(pair[1])
			if err != nil {
				return dbi, err
			}
			dbi.DBPort = port
		} else {
			dbi.DBHost = hp
		}

		if uri.User != nil {
			dbi.DBUser = uri.User.Username()
			if pass, ok := uri.User.Password(); ok {
				dbi.DBPassword = pass
			}
		}

		p := uri.Path
		if p != "" && p != "/" {
			if strings.ContainsAny(p, "/") {
				p = p[1:]
			}
			dbi.DBName = p
		}
	}
	return dbi, nil
}

func (dbi *DBInfo) FillEmptyFromURI() error {
	tmp, err := parseDBURI(dbi.DBURI)
	if err != nil {
		return err
	}
	if dbi.DBHost == "" {
		dbi.DBHost = tmp.DBHost
	}
	if dbi.DBPort == 0 {
		dbi.DBPort = tmp.DBPort
	}
	if dbi.DBName == "" {
		dbi.DBName = tmp.DBName
	}
	if dbi.DBUser == "" {
		dbi.DBUser = tmp.DBUser
	}
	if dbi.DBPassword == "" {
		dbi.DBPassword = tmp.DBPassword
	}
	return nil
}

type Config struct {
	DBInfo
	MediaExts   []string `json:"media_exts"`
	Debug       bool     `json:"debug"`
	Development bool     `json:"development"`
	AuthSecret  string   `json:"auth_secret"`
	AuthTimeout Duration `json:"auth_timeout"`
}

func LoadConfig(filename string) (*Config, error) {

	cfg := &Config{}
	file, err := os.Open(filename)

	if err != nil {
		return cfg, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return cfg, err
	}
	err = cfg.FillEmptyFromURI()
	return cfg, err

}
