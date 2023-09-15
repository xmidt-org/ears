// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	log.Logger = log.With().Str("app.id", "ears").Logger()
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
