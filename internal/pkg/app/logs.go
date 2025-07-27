// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
)

func ProvideLogger(config config.Config) (*zerolog.Logger, error) {
	logLevel, err := zerolog.ParseLevel(config.GetString("ears.logLevel"))
	if err != nil {
		return nil, &InvalidOptionError{
			Option: fmt.Sprintf("loglevel %s is not valid", config.GetString("ears.logLevel")),
		}
	}
	hostname := config.GetString("ears.hostname")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	logger := zerolog.New(os.Stdout).Level(logLevel).With().
		Str(rtsemconv.EarsLogHostnameKey, hostname).
		Timestamp().Logger()
	zerolog.LevelFieldName = "log.level"
	return &logger, nil
}
