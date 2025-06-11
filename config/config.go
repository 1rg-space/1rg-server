package config

import "github.com/BurntSushi/toml"

var Config struct {
	Cals struct {
		SecondFloor  string `toml:"second_floor"`
		GreenRoom    string `toml:"green_room"`
		PurpleRoom   string `toml:"purple_room"`
		PublicEvents string `toml:"public_events"`
	} `toml:"cals"`
	DBPath       string `toml:"db_path"`
	AssetStorage string `toml:"asset_storage"`
}

func LoadConfig() error {
	_, err := toml.DecodeFile("config.toml", &Config)
	return err
}
