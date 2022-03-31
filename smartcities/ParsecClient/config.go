package main

import (
	"github.com/spf13/viper"
)

type CfgApp struct {
	Port int
}

type CfgLog struct {
	Level      string
	FileName   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

type AppConfig struct {
	App CfgApp `mapstructure:"app"`
	Log CfgLog `mapstructure:"log"`
}

var Conf AppConfig

func InitConfig() {
	viper.SetConfigName("ParsecClient") // config file name
	viper.SetConfigType("toml")         // config file ext
	viper.AddConfigPath(".")            // config file find path
	viper.AddConfigPath("/etc")         // config file find path
	err := viper.ReadInConfig()         // find and read config
	if err != nil {
		panic(err)
	}

	// parse config
	if err := viper.Unmarshal(&Conf); err != nil {
		panic(err)
	}
}
