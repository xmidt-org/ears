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
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/app"
	"github.com/xmidt-org/ears/internal/pkg/appsecret"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/fx/checkpointmanagerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/fragmentstorerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/jwtmanagerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/nodestatemanagerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/pluginmanagerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/quotamanagerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/routestorerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/syncerfx"
	"github.com/xmidt-org/ears/internal/pkg/fx/tenantstorerfx"
	"github.com/xmidt-org/ears/internal/pkg/tablemgr"
	"github.com/xmidt-org/ears/pkg/cli"
	"github.com/xmidt-org/ears/pkg/panics"
	"go.uber.org/fx"
	"os"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the EARS microservice",
	Long:  `Runs the EARS microservice`,
	Run: func(cmd *cobra.Command, args []string) {

		defer func() {
			p := recover()
			if p != nil {
				panicErr := panics.ToError(p)
				log.Logger.Fatal().Str("op", "run").Str("error", panicErr.Error()).
					Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred during startup")
			}
		}()

		//create an init logger for uberfx initialization
		//Uberfx uses Prtinf to log initialization messages, but
		//zerolog only logs at debug level when using Printf. If
		//we don't use a separate zerolog logger, the error
		//message is suppressed if logLevel is above debug
		initLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
		zerolog.LevelFieldName = "log.level"

		earsApp := fx.New(
			pluginmanagerfx.Module,
			routestorerfx.Module,
			fragmentstorerfx.Module,
			syncerfx.Module,
			tenantstorerfx.Module,
			quotamanagerfx.Module,
			nodestatemanagerfx.Module,
			checkpointmanagerfx.Module,
			jwtmanagerfx.Module,
			fx.Provide(
				AppConfig,
				appsecret.NewConfigVault,
				app.ProvideLogger,
				tablemgr.NewRoutingTableManager,
				app.NewAPIManager,
				app.NewMiddleware,
				app.NewMux,
			),
			fx.Logger(&initLogger),
			fx.Invoke(syncerfx.SetupDeltaSyncer),
			fx.Invoke(tablemgr.SetupRoutingManager),
			fx.Invoke(quotamanagerfx.SetupQuotaManager),
			fx.Invoke(app.SetupOpenTelemetry),
			fx.Invoke(app.SetupAPIServer),
			fx.Invoke(app.SetupPprof),
			fx.Invoke(app.SetupNodeStateManager),
			fx.Invoke(app.SetupCheckpointManager),
		)
		earsApp.Run()
	},
}

func AppConfig() config.Config {
	return viper.GetViper()
}

func init() {
	rootCmd.AddCommand(runCmd)
	cli.ViperAddArguments(
		runCmd,
		[]cli.Argument{
			cli.Argument{
				Name: "logLevel", Shorthand: "", Type: cli.ArgTypeString,
				Default: "info", LookupKey: "ears.logLevel",
				Description: "log level",
			},
			cli.Argument{
				Name: "port", Shorthand: "", Type: cli.ArgTypeInt,
				Default: 8080, LookupKey: "ears.api.port",
				Description: "API port",
			},
			cli.Argument{
				Name: "routeStorageType", Shorthand: "", Type: cli.ArgTypeString,
				Default: "inmemory", LookupKey: "ears.storage.route.type",
				Description: "persistence layer storage type for routes (inmemory, dynamodb)",
			},
			cli.Argument{
				Name: "routeStorageDynamoRegion", Shorthand: "", Type: cli.ArgTypeString,
				Default: "us-west-2", LookupKey: "ears.storage.route.region",
				Description: "route dynamodb region",
			},
			cli.Argument{
				Name: "routeStorageDynamoTable", Shorthand: "", Type: cli.ArgTypeString,
				Default: "dev.ears.routes", LookupKey: "ears.storage.route.table",
				Description: "route dynamodb table name",
			},
			cli.Argument{
				Name: "tenantStorageType", Shorthand: "", Type: cli.ArgTypeString,
				Default: "inmemory", LookupKey: "ears.storage.tenant.type",
				Description: "persistence layer storage type for tenants (inmemory, dynamodb)",
			},
			cli.Argument{
				Name: "tenantStorageDynamoRegion", Shorthand: "", Type: cli.ArgTypeString,
				Default: "us-west-2", LookupKey: "ears.storage.tenant.region",
				Description: "tenant dynamodb region",
			},
			cli.Argument{
				Name: "tenantStorageDynamoTable", Shorthand: "", Type: cli.ArgTypeString,
				Default: "dev.ears.tenant", LookupKey: "ears.storage.tenant.table",
				Description: "tenant dynamodb table name",
			},
			cli.Argument{
				Name: "redisEndpoint", Shorthand: "", Type: cli.ArgTypeString,
				Default: "gears-redis-qa-001.6bteey.0001.usw2.cache.amazonaws.com:6379", LookupKey: "ears.synchronization.endpoint",
				Description: "redis endpoint for routing table synchronization",
			},
		},
	)
}
