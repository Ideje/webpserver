package config

import (
	"github.com/spf13/viper"
)

var (
	Cfg *Config
)

type Config struct {
	ServerAddress   string `addr:"server address"`
	ServerName      string `servername:"server name"`
	ImageSrcDir     string `imagesrcdir:"image source directory"`
	ImageDstDir     string `imagedstdir:"image destination directory"`
	LogLevel        string `loglevel:"log level"`
	WebPMaxAge      int    `webpmaxage:"WebP max-age cache control header"`
	HeaderCacheInfo bool   `headercacheinfo:"header cache info"`
}

func LoadConfig(path string) (err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("webp")
	viper.SetConfigType("env")

	viper.SetDefault("ServerAddress", ":http")

	viper.SetEnvPrefix("webp")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&Cfg)
	return
}
