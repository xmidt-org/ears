package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/xmidt-org/ears/pkg/app"
	"github.com/xmidt-org/ears/pkg/cli"
	"os"
	"os/signal"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the EARS microservice",
	Long:  `Runs the EARS microservice`,
	Run: func(cmd *cobra.Command, args []string) {

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-c
			cancel()
		}()

		if err := app.Run(ctx); nil != err {
			//not log the error since the throwing error func already did
			os.Exit(1)
		}
	},
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
