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

package cli_test

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/pkg/cli"
	"reflect"
	"testing"
)

func TestConfigErrors(t *testing.T) {

	cmd := &cobra.Command{
		Use:   "use",
		Short: "short",
		Long:  `long`,
	}

	cli.ViperAddArgument(
		cmd,
		cli.Argument{
			Name: "config", Shorthand: "", Type: cli.ArgTypeString,
			Default: "", Persistent: true,
			Description: "config file (default is $HOME/cli.yaml)",
		},
	)

	testCases := [][]string{
		[]string{"noextension", `ConfigError (path=noextension): open noextension: no such file or directory`},
		[]string{"./mymissingfile.yaml", `ConfigError (path=./mymissingfile.yaml): open ./mymissingfile.yaml: no such file or directory`},
		[]string{"./config.weirdext", `ConfigError (path=./config.weirdext): Unsupported Config Type "weirdext"`},
		[]string{"file://config", `ConfigError (path=config): open config: no such file or directory`},
		[]string{"file://config.yaml", `ConfigError (path=config.yaml): open config.yaml: no such file or directory`},
		[]string{"f://random/protocol.json", `ConfigNotSupportedProtocolError (protocol=f)`},

		// TODO:  Need to look into possibly not embedding dynamic info into the error string
		// such as request and host ids (which the aws error does)
		// []string{"s3://clibucketthatdoesntexist/file.yaml", `ConfigNotSupportedProtocol, path: s3://clibucketthatdoesntexist/file.yaml`},
		// Result:
		//   ConfigProcess: get_object: NoSuchBucket: The specified bucket does not exist 	status code: 404, request id: 35CE530C839CFF6E, host id: doAcXEzm0LXXEHr0APmYG+T9JlnFDaN4HIMvnwpzIZNJvRm8+FqqrL5dkC6VfesJ0QfNPKFsUQo=, url: s3://clibucketthatdoesntexist/file.yaml, path: s3://clibucketthatdoesntexist/file.yaml

	}

	for i, test := range testCases {
		viper.Set("config", test[0])

		err := cli.ViperConfig("", "")

		if err == nil {
			t.Errorf("expected '%s' but error was <nil>", test[1])
		} else {
			if err.Error() != test[1] {
				t.Errorf("#%d expected '%s' != '%s'", i, test[1], err.Error())
			}
		}
	}
}

func TestConfigError(t *testing.T) {
	err := cli.ViperAddArgument(
		nil,
		cli.Argument{
			Name: "config", Shorthand: "", Type: cli.ArgTypeString,
			Default: "", Persistent: true,
			Description: "config file (default is $HOME/cli.yaml)",
		},
	)
	var emptyCmdErr *cli.EmptyCmdArgumentError
	if err == nil {
		t.Errorf("Expect an error, got no error")
		return
	}
	if !errors.As(err, &emptyCmdErr) {
		t.Errorf("Expect ArgError type, got a different error type %s", reflect.TypeOf(err))
	}

	cmd := &cobra.Command{
		Use:   "use",
		Short: "short",
		Long:  `long`,
	}
	err = cli.ViperAddArgument(
		cmd,
		cli.Argument{
			Name: "config", Shorthand: "", Type: cli.ArgTypeString,
			Default: nil, Persistent: true,
			Description: "config file (default is $HOME/cli.yaml)",
		},
	)
	var noDefaultSetErr *cli.NoDefaultSetError
	if err == nil {
		t.Errorf("Expect an error, got no error")
		return
	}
	if !errors.As(err, &noDefaultSetErr) {
		t.Errorf("Expect ArgError type, got a different error type %s", reflect.TypeOf(err))
	}
}

func TestConfigRead(t *testing.T) {

	cmd := &cobra.Command{
		Use:   "use",
		Short: "short",
		Long:  `long`,
	}

	cli.ViperAddArgument(
		cmd,
		cli.Argument{
			Name: "config", Shorthand: "", Type: cli.ArgTypeString,
			Default: "", Persistent: true,
			Description: "config file (default is $HOME/cli.yaml)",
		},
	)

	cli.ViperAddArguments(
		cmd,
		[]cli.Argument{
			cli.Argument{
				Name: "string", Shorthand: "", Type: cli.ArgTypeString,
				Default: "", LookupKey: "string",
				Description: "A string",
			},
			cli.Argument{
				Name: "boolean", Shorthand: "", Type: cli.ArgTypeBool,
				Default: false, LookupKey: "boolean",
				Description: "A bool",
			},
			cli.Argument{
				Name: "number", Shorthand: "", Type: cli.ArgTypeInt,
				Default: 8080, LookupKey: "number",
				Description: "An int",
			},
			cli.Argument{
				Name: "array", Shorthand: "", Type: cli.ArgTypeStringSlice,
				Default: []string{}, LookupKey: "array",
				Description: "A string array",
			},
		},
	)

	testCases := []struct {
		file   string
		values map[string]interface{}
	}{
		{
			file: "file://./testdata/cli.yaml",
			values: map[string]interface{}{
				"string":  "Here is a string",
				"boolean": true,
				"number":  12,
				"array":   []interface{}{"one", "two", "three"},
			},
		},
	}

	for _, test := range testCases {

		viper.Set("config", test.file)
		cli.ViperConfig("", "")

		for k, v := range test.values {
			value := viper.Get(k)
			if !cmp.Equal(v, value) {
				t.Errorf("expected '%v' != '%v'", v, value)
			}
		}
	}
}
