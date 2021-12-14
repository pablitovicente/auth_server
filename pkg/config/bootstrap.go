package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Bootstrap(filePath string) *viper.Viper {

	config := viper.New()
	config.SetConfigType("json")
	config.SetConfigName("config")
	config.AddConfigPath(filePath)

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("config error: %w", err))
	}

	return config
}
