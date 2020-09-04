/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package app

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var Version = "v0.0.0"

type App struct {
	ctx context.Context
}

func initLogging() error {
	//Initialize logging further
	logLevel, err := zerolog.ParseLevel(viper.GetString("ears.logLevel"))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.With().Str("env", viper.GetString("ears.env")).Logger()
	return nil
}

//Run is called after Viper/Cobra commands setup
func Run() error {
	err := initLogging()
	if err != nil {
		log.Error().Str("op", "initApp").Msg(err.Error())
		return err
	}
	port := viper.GetInt("ears.api.port")
	err = startAPIServer(port)
	return nil
}
