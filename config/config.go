package config

import "github.com/BurntSushi/toml"

var Config struct {
	Cals struct {
		SecondFloor string `toml:"second_floor"`
		GreenRoom   string `toml:"green_room"`
		PurpleRoom  string `toml:"purple_room"`
	} `toml:"cals"`
}

func LoadConfig() error {
	_, err := toml.DecodeFile("config.toml", &Config)
	return err
}
