package main

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/dwarfops/lgtm-auth-proxy/cmd"
	"github.com/dwarfops/lgtm-auth-proxy/utils"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/lgtm-auth-proxy")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msgf("Failed to load config")
	}
	logLevel := viper.GetString("log_level")
	if err := utils.SetZerologLevel(logLevel); err != nil {
		log.Fatal().Err(err).Msgf("Failed to set log level: '%s'", logLevel)
	}
}

func main() {
	if err := cmd.Root.ExecuteContext(context.Background()); err != nil {
		log.Fatal().Err(err).Msgf("Failed to execute cmd.")
	}
}
