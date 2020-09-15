package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"

	"github.com/rs/zerolog/log"
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

		logger, err := InitGlobalLogger(ViperConfig())
		if err != nil {
			log.Logger.Error().Str("op", "InitLogger").Msg(err.Error())
			os.Exit(1)
		}
		earsApp := fx.New(
			fx.Provide(
				ViperConfig,
				app.NewLogger,
				app.NewRouter,
				app.NewMiddlewares,
				app.NewMux,
			),
			fx.Logger(logger),
			fx.Invoke(app.SetupAPIServer),
		)
		earsApp.Run()
	},
}

func ViperConfig() app.Config {
	return viper.Sub("ears")
}

//Initialize logging. We are initializing the zerolog global logger for uber/fx logging
//The application itself will not use the global logger. Instead, it use the logger returned by
//app.NewLogger function, which is a child logger of the global logger
func InitGlobalLogger(config app.Config) (fx.Printer, error) {
	logLevel, err := zerolog.ParseLevel(config.GetString("logLevel"))
	if err != nil {
		return nil, &app.InvalidOptionError{
			Err: fmt.Errorf("invalid loglevel %s error %s",
				config.GetString("logLevel"),
				err.Error()),
		}
	}
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.With().Str("env", config.GetString("env")).Logger()
	return &log.Logger, nil
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
