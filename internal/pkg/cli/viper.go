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

package cli

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

// ViperAddArgument provides a data driven way to add cli arguments
func ViperAddArgument(cmd *cobra.Command, a Argument) error {

	if cmd == nil {
		return &ArgError{errors.New("nil cmd argument"), "cmd", nil}
	}

	if a.Default == nil {
		return &ArgError{errors.New("no default set"), "a", a}
	}

	var fs *pflag.FlagSet

	if a.Persistent {
		fs = cmd.PersistentFlags()
	} else {
		fs = cmd.Flags()
	}

	// TODO: Fill this in with more types
	switch a.Type {
	case ArgTypeBool:
		fs.BoolP(a.Name, a.Shorthand, a.Default.(bool), a.Description)
	case ArgTypeBoolSlice:
		fs.BoolSliceP(a.Name, a.Shorthand, a.Default.([]bool), a.Description)
	case ArgTypeInt:
		fs.IntP(a.Name, a.Shorthand, a.Default.(int), a.Description)
	case ArgTypeIntSlice:
		fs.IntSliceP(a.Name, a.Shorthand, a.Default.([]int), a.Description)
	case ArgTypeString:
		fs.StringP(a.Name, a.Shorthand, a.Default.(string), a.Description)
	case ArgTypeStringSlice:
		fs.StringSliceP(a.Name, a.Shorthand, a.Default.([]string), a.Description)
	}

	if a.LookupKey == "" {
		a.LookupKey = a.Name
	}
	viper.BindPFlag(a.LookupKey, fs.Lookup(a.Name))

	return nil
}

// ViperAddArguments allows you to pass in many argument configs
func ViperAddArguments(cmd *cobra.Command, aList []Argument) error {
	if cmd == nil {
		return &ArgError{errors.New("nil cmd argument"), "cmd", nil}
	}

	for _, a := range aList {
		err := ViperAddArgument(cmd, a)
		if err != nil {
			return err
		}
	}

	return nil
}

// ViperConfig will help configure a project. This includes setting
// the environment variable prefix, plus the default config file name.
// For example:
//   cli.ViperConfig("ears")
func ViperConfig(configName string) error {

	viper.SetConfigName(configName)

	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// NOTE: This needs to go after setting the config name and paths
	// Otherwise, the config name + paths will override this setting

	// If a config file is found, read it in.
	config := viper.GetString("config")

	if config == "" {
		err := viper.ReadInConfig()

		if err != nil {
			// The most common case is that the default config file name will not
			// be found.  If this is the case, we will ignore this error.
			var fileNotFoundErr *viper.ConfigFileNotFoundError
			if err != nil && !errors.As(err, &fileNotFoundErr) {
				return &ConfigError{err, config}
			}
		}
		return nil
	}

	// If any error happens with reading in this config file, we throw
	// an error.  The expectation it that the file should exist and
	// be valid if it's being passed in.

	ext := strings.TrimLeft(filepath.Ext(config), ".")
	if ext == "" {
		ext = "yaml"
	}
	viper.SetConfigType(ext)

	parts := strings.SplitN(config, "://", 2)
	switch parts[0] {
	case "http", "https":
		//TODO: do we need this?

	case "s3":
		//TODO: do we need this?

	case "file":
		// Set parts[0] to the file path and fall through to file path processing
		config = parts[1]
		parts = []string{}
		fallthrough
	default:
		// Value is a path to a local file
		if len(parts) > 1 {
			return &ConfigError{errors.New(ErrConfigNotSupportedProtocol), config}
		}
		viper.SetConfigFile(config)
		if err := viper.ReadInConfig(); err != nil {
			// It was intended to load in a config and we could not.
			// Return an error
			return &ConfigError{err, config}
		}

	}

	return nil
}
