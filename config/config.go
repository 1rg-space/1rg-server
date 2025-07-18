package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Build-time variable
var isProdStr = "false"

var IsProduction bool

var Config struct {
	Cals struct {
		SecondFloor string `toml:"second_floor"`
		GreenRoom   string `toml:"green_room"`
		PurpleRoom  string `toml:"purple_room"`
		Events      string `toml:"events"`
	} `toml:"cals"`
	DBPath       string `toml:"db_path"`
	AssetStorage string `toml:"asset_storage"`
	CSRFKey      string `toml:"csrf_key"`
}

func LoadConfig(path string) error {
	if isProdStr == "true" {
		IsProduction = true
	}

	_, err := toml.DecodeFile(path, &Config)
	if err != nil {
		return err
	}
	if Config.DBPath == "" {
		return fmt.Errorf("must specify db_path in config")
	}
	if Config.AssetStorage == "" {
		return fmt.Errorf("must specify asset_storage in config")
	}
	if Config.CSRFKey == "" {
		return fmt.Errorf("must specify csrf_key in config")
	}
	return nil
}

func CalendarsProvided() bool {
	if Config.Cals.SecondFloor == "" {
		return false
	}
	if Config.Cals.GreenRoom == "" {
		return false
	}
	if Config.Cals.PurpleRoom == "" {
		return false
	}
	if Config.Cals.Events == "" {
		return false
	}
	return true
}
