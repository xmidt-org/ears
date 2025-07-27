// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package subcmd

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/xmidt-org/ears/pkg/cli"
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
	log.Logger = log.With().Str("service.name", "ears").Logger()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Str("op", "Execute").Msg(err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cli.ViperAddArguments(
		rootCmd,
		[]cli.Argument{
			cli.Argument{
				Name: "config", Shorthand: "", Type: cli.ArgTypeString,
				Default: "", Persistent: true,
				Description: "config file (default is $HOME/ears.yaml)",
			},
		},
	)
}

func initConfig() {
	err := cli.ViperConfig("EARS", "ears")
	if err != nil {
		log.Fatal().Str("op", "initConfig").Msg(err.Error())
	}
}
