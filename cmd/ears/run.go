package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/app"
	"github.com/xmidt-org/ears/internal/pkg/cli"
	"go.uber.org/fx"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the EARS microservice",
	Long:  `Runs the EARS microservice`,
	Run: func(cmd *cobra.Command, args []string) {

		earsApp := fx.New(
			fx.Provide(
				ViperConfig,
				app.NewLogger,
				app.NewRouter,
				app.NewMiddlewares,
				app.NewMux,
			),
			fx.Invoke(app.SetupAPIServer),
		)
		earsApp.Run()
	},
}

func ViperConfig() app.Config {
	return viper.Sub("ears")
}

func init() {
	rootCmd.AddCommand(runCmd)
	cli.ViperAddArguments(
		runCmd,
		[]cli.Argument{
			cli.Argument{
				Name: "env", Shorthand: "", Type: cli.ArgTypeString,
				Default: "local", LookupKey: "ears.env",
				Description: "Environment",
			},
			cli.Argument{
				Name: "logLevel", Shorthand: "", Type: cli.ArgTypeString,
				Default: "info", LookupKey: "ears.debugLevel",
				Description: "log level",
			},
			cli.Argument{
				Name: "port", Shorthand: "", Type: cli.ArgTypeInt,
				Default: 8080, LookupKey: "ears.api.port",
				Description: "API port",
			},
		},
	)
}
