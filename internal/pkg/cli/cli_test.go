package cli_test

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/cli"
	"reflect"
	"strings"
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
		[]string{"noextension", `open noextension: no such file or directory, path: noextension`},
		[]string{"./mymissingfile.yaml", `open ./mymissingfile.yaml: no such file or directory, path: ./mymissingfile.yaml`},
		[]string{"./config.weirdext", `Unsupported Config Type "weirdext", path: ./config.weirdext`},
		[]string{"file://config", `open config: no such file or directory, path: config`},
		[]string{"file://config.yaml", `open config.yaml: no such file or directory, path: config.yaml`},
		[]string{"f://random/protocol.json", `ConfigNotSupportedProtocol, path: f://random/protocol.json`},

		// TODO:  Need to look into possibly not embedding dynamic info into the error string
		// such as request and host ids (which the aws error does)
		// []string{"s3://clibucketthatdoesntexist/file.yaml", `ConfigNotSupportedProtocol, path: s3://clibucketthatdoesntexist/file.yaml`},
		// Result:
		//   ConfigProcess: get_object: NoSuchBucket: The specified bucket does not exist 	status code: 404, request id: 35CE530C839CFF6E, host id: doAcXEzm0LXXEHr0APmYG+T9JlnFDaN4HIMvnwpzIZNJvRm8+FqqrL5dkC6VfesJ0QfNPKFsUQo=, url: s3://clibucketthatdoesntexist/file.yaml, path: s3://clibucketthatdoesntexist/file.yaml

	}

	for i, test := range testCases {
		viper.Set("config", test[0])

		err := cli.ViperConfig("")

		switch err {
		case nil:
			t.Errorf("expected '%s' but error was <nil>", test[1])
		default:
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
	var argErr *cli.ArgError
	if err == nil {
		t.Errorf("Expect an error, got no error")
		return
	}
	if !errors.As(err, &argErr) {
		t.Errorf("Expect ArgError type, got a different error type %s", reflect.TypeOf(err))
	}
	argErr, _ = err.(*cli.ArgError)
	if argErr.Unwrap().Error() != "nil cmd argument" {
		t.Errorf("Unexpected error message %s", argErr.Unwrap().Error())
	}
	if err.Error() != "nil cmd argument, key=cmd" {
		t.Errorf("Unexpected error message %s", err.Error())
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
	if err == nil {
		t.Errorf("Expect an error, got no error")
		return
	}
	if !errors.As(err, &argErr) {
		t.Errorf("Expect ArgError type, got a different error type %s", reflect.TypeOf(err))
	}
	argErr, _ = err.(*cli.ArgError)
	if argErr.Unwrap().Error() != "no default set" {
		t.Errorf("Unexpected error message %s", argErr.Unwrap().Error())
	}
	if !strings.HasPrefix(err.Error(), "no default set, key=a, value=") {
		t.Errorf("Unexpected error message %s", err.Error())
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
		cli.ViperConfig("")

		for k, v := range test.values {
			value := viper.Get(k)
			if !cmp.Equal(v, value) {
				t.Errorf("expected '%v' != '%v'", v, value)
			}
		}
	}
}
