package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/xmidt-org/ears/internal/pkg/cli"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ears",
	Short: "The EARS microservice",
	Long:  `The Event Async Routing Service`,
}

func Execute() {
	//Initialize logging for command setup. Log level/env will be set
	//later when we read in the configurations
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Str("app.id", "ears").Logger()

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Str("op", "Execute").Msg(err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	err := cli.ViperConfig("ears")
	if err != nil {
		log.Fatal().Str("op", "initConfig").Msg(err.Error())
	}
}
