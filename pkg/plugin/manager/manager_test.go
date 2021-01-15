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

package manager_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/bit"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/plugin/manager"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"

	. "github.com/onsi/gomega"
)

const (
	buildPluginTimeout = 1 * time.Minute

	testPluginDir = "testplugins"
)

var logger *log.Logger

func TestMain(m *testing.M) {

	flag.Parse()

	if verboseEnabled() {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	} else {
		logger = log.New(ioutil.Discard, "", log.LstdFlags)
	}

	shutdown := func(code int) {
		err := testShutdown()
		if err != nil {
			logger.Println(err)
			os.Exit(1 | code)
		}
		os.Exit(code)
	}

	err := testSetup()
	if err != nil {
		logger.Println(err)
		shutdown(1)
	}

	shutdown(m.Run())
}

func TestNilPluginError(t *testing.T) {
	testCases := []struct {
		name string
		plug plugin.Pluginer
		err  error
	}{
		{name: "nil_plugin", err: &manager.NilPluginError{}},
	}

	m, _ := manager.New()

	for _, tc := range testCases {
		t.Run("register_"+tc.name, func(t *testing.T) {
			a := NewWithT(t)
			err := m.RegisterPlugin(tc.name, tc.plug)
			if tc.err == nil {
				a.Expect(err).To(BeNil())
			} else {
				a.Expect(err).To(Equal(tc.err))
			}

		})
	}

}

func TestAlreadyRegisteredError(t *testing.T) {
	testCases := []struct {
		name string
		plug plugin.Pluginer
		err  error
	}{
		{name: "registered", plug: &newPluginererMock{}},
		{name: "registered", plug: &newPluginererMock{}, err: &manager.AlreadyRegisteredError{}},
	}

	m, _ := manager.New()

	for _, tc := range testCases {
		t.Run("register_"+tc.name, func(t *testing.T) {
			a := NewWithT(t)
			err := m.RegisterPlugin(tc.name, tc.plug)
			if tc.err == nil {
				a.Expect(err).To(BeNil())
			} else {
				a.Expect(err).To(Equal(tc.err))
			}

		})
	}

}

func TestRegistrationLifecycle(t *testing.T) {
	testCases := []struct {
		name string
		plug plugin.Pluginer
	}{
		{name: "one", plug: &newPluginererMock{}},
		{name: "two", plug: &newPluginererMock{}},
		{name: "three", plug: &newPluginererMock{}},
		{name: "four", plug: &newPluginererMock{}},
	}

	m, _ := manager.New()
	a := NewWithT(t)

	for _, tc := range testCases {
		t.Run("register_"+tc.name, func(t *testing.T) {
			a := NewWithT(t)
			err := m.RegisterPlugin(tc.name, tc.plug)
			a.Expect(err).To(BeNil())

		})
	}

	registrations := m.Plugins()
	a.Expect(len(testCases)).To(Equal(len(registrations)))

	for _, tc := range testCases {
		t.Run("lookup_"+tc.name, func(t *testing.T) {
			a := NewWithT(t)
			_, ok := registrations[tc.name]
			a.Expect(ok).To(BeTrue())

			r := m.Plugin(tc.name)
			a.Expect(r).ToNot(BeNil())

		})
	}

	for _, tc := range testCases {
		t.Run("unregister_"+tc.name, func(t *testing.T) {
			a := NewWithT(t)
			err := m.UnregisterPlugin(tc.name)
			a.Expect(err).To(BeNil())
		})
	}

	registrations = m.Plugins()
	a.Expect(len(registrations)).To(Equal(0))

}

func TestRegistrationTypes(t *testing.T) {

	testCases := []struct {
		name         string
		numPlugins   int
		numReceivers int
		numSenders   int
		numFilterers int
	}{
		{
			name:         "only_plugins",
			numPlugins:   3,
			numReceivers: 0,
			numSenders:   0,
			numFilterers: 0,
		},
		{
			name:         "only_receivers",
			numPlugins:   0,
			numReceivers: 5,
			numSenders:   0,
			numFilterers: 0,
		},
		{
			name:         "only_senders",
			numPlugins:   0,
			numReceivers: 0,
			numSenders:   2,
			numFilterers: 0,
		},
		{
			name:         "only_filters",
			numPlugins:   0,
			numReceivers: 0,
			numSenders:   0,
			numFilterers: 4,
		},
		{
			name:         "mix",
			numPlugins:   3,
			numReceivers: 5,
			numSenders:   2,
			numFilterers: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			a := NewWithT(t)
			m, _ := manager.New()

			// Plugins
			for i := 0; i < tc.numPlugins; i++ {
				err := m.RegisterPlugin(
					"plugins_"+strconv.Itoa(i),
					&newPluginererMock{},
				)
				a.Expect(err).To(BeNil())
			}

			// Receivers
			for i := 0; i < tc.numReceivers; i++ {
				err := m.RegisterPlugin(
					"receivers_"+strconv.Itoa(i),
					&newReceivererMock{},
				)
				a.Expect(err).To(BeNil())
			}

			// Fiterers
			for i := 0; i < tc.numFilterers; i++ {
				err := m.RegisterPlugin(
					"filterers_"+strconv.Itoa(i),
					&newFiltererMock{},
				)
				a.Expect(err).To(BeNil())
			}

			// Senders
			for i := 0; i < tc.numSenders; i++ {
				err := m.RegisterPlugin(
					"senders_"+strconv.Itoa(i),
					&newSendererMock{},
				)
				a.Expect(err).To(BeNil())
			}

			receiverers := m.Receiverers()
			a.Expect(len(receiverers)).To(Equal(tc.numReceivers))

			filterers := m.Filterers()
			a.Expect(len(filterers)).To(Equal(tc.numFilterers))

			senderers := m.Senderers()
			a.Expect(len(senderers)).To(Equal(tc.numSenders))

			registrations := m.Plugins()
			a.Expect((len(registrations))).To(Equal(tc.numPlugins + tc.numReceivers + tc.numFilterers + tc.numSenders))

		})
	}

}

func TestLoadPlugin(t *testing.T) {

	paths, err := getTestPluginSOPaths()
	if err != nil {
		t.Errorf("could not get plugin paths: %w", err)
	}

	m, _ := manager.New()
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			a := NewWithT(t)

			p, err := m.LoadPlugin(
				manager.Config{
					Name: path,
					Path: path,
				},
			)

			var expectedErr error
			if strings.Contains(path, "/err_") {
				switch {
				case strings.Contains(path, "NewPluginerNotImplementedError"):
					expectedErr = &manager.NewPluginerNotImplementedError{}
				case strings.Contains(path, "VariableLookupError"):
					expectedErr = &manager.VariableLookupError{}
				case strings.Contains(path, "NewPluginerError"):
					expectedErr = &manager.NewPluginerError{}
				}
			}

			if expectedErr == nil {
				a.Expect(err).To(BeNil())
				a.Expect(p).ToNot(BeNil())

				g := goldie.New(t, goldie.WithTestNameForDir(true))
				r := m.Plugin(path)
				a.Expect(r).ToNot(BeNil())

				g.AssertJson(t, strings.ReplaceAll(path, "/", "_"), r.Capabilities)

			} else {
				a.Expect(p).To(BeNil())
				a.Expect(errTypeToString(err)).To(Equal(errTypeToString(expectedErr)))
			}

		})
	}
}

func TestLoadErrors(t *testing.T) {

	testCases := []struct {
		name   string
		config manager.Config
		err    error
	}{
		{
			name:   "missing_config_name",
			config: manager.Config{},
			err:    &manager.InvalidConfigError{Err: fmt.Errorf("config name cannot be empty")},
		},
		{
			name:   "missing_config_path",
			config: manager.Config{Name: "missing_config_path"},
			err:    &manager.InvalidConfigError{Err: fmt.Errorf("config path cannot be empty")},
		},
		{
			name:   "bad_plugin_path",
			config: manager.Config{Name: "bad_plugin_path", Path: "bad_plugin_path"},
			err:    &manager.OpenPluginError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)
			m, _ := manager.New()

			_, err := m.LoadPlugin(tc.config)
			a.Expect(err).ToNot(BeNil())
			a.Expect(errTypeToString(err)).To(Equal(errTypeToString(tc.err)))
		})
	}

}

func TestNewReceiver(t *testing.T) {

	a := NewWithT(t)
	m, _ := manager.New()

	nfe := &manager.NotFoundError{}

	{
		p, err := m.NewReceiver("myreceiver", "")
		a.Expect(p).To(BeNil())
		a.Expect(errTypeToString(err)).To(Equal(errTypeToString(nfe)))
	}

	{
		m.RegisterPlugin(
			"myreceiver",
			&newReceivererMock{
				receiver.NewReceivererMock{
					NewReceiverFunc: func(config interface{}) (receiver.Receiver, error) {
						return &receiver.ReceiverMock{}, nil
					},
				},
			},
		)
		p, err := m.NewReceiver("myreceiver", "")
		a.Expect(p).ToNot(BeNil())
		a.Expect(err).To(BeNil())
		m.UnregisterPlugin("myreceiver")
	}

	{
		p, err := m.NewReceiver("myreceiver", "")
		a.Expect(p).To(BeNil())
		a.Expect(errTypeToString(err)).To(Equal(errTypeToString(nfe)))
	}

}

func TestNewFilterer(t *testing.T) {
	a := NewWithT(t)
	m, _ := manager.New()

	nfe := &manager.NotFoundError{}

	{
		p, err := m.NewFilterer("myfilterer", "")
		a.Expect(p).To(BeNil())
		a.Expect(errTypeToString(err)).To(Equal(errTypeToString(nfe)))
	}

	{
		m.RegisterPlugin(
			"myfilterer",
			&newFiltererMock{
				filter.NewFiltererMock{
					NewFiltererFunc: func(config interface{}) (filter.Filterer, error) {
						return &filter.FiltererMock{}, nil
					},
				},
			},
		)
		p, err := m.NewFilterer("myfilterer", "")
		a.Expect(p).ToNot(BeNil())
		a.Expect(err).To(BeNil())
		m.UnregisterPlugin("myfilterer")
	}

	{
		p, err := m.NewFilterer("myfilterer", "")
		a.Expect(p).To(BeNil())
		a.Expect(errTypeToString(err)).To(Equal(errTypeToString(nfe)))
	}

}

func TestNewSenderer(t *testing.T) {
	a := NewWithT(t)
	m, _ := manager.New()

	nfe := &manager.NotFoundError{}

	{
		p, err := m.NewSender("mysenderer", "")
		a.Expect(p).To(BeNil())
		a.Expect(errTypeToString(err)).To(Equal(errTypeToString(nfe)))
	}

	{
		m.RegisterPlugin(
			"mysenderer",
			&newSendererMock{
				sender.NewSendererMock{
					NewSenderFunc: func(config interface{}) (sender.Sender, error) {
						return &sender.SenderMock{}, nil
					},
				},
			},
		)
		p, err := m.NewSender("mysenderer", "")
		a.Expect(p).ToNot(BeNil())
		a.Expect(err).To(BeNil())
		m.UnregisterPlugin("mysenderer")
	}

	{
		p, err := m.NewSender("mysenderer", "")
		a.Expect(p).To(BeNil())
		a.Expect(errTypeToString(err)).To(Equal(errTypeToString(nfe)))
	}

}

// === Test Lifecycle ===========================

func testSetup() error {
	return buildTestPlugins()
}

func testShutdown() error {
	files, err := getTestPluginSOPaths()

	if err != nil {
		logger.Printf("error cleaning up *.so: %s", err.Error())
	}

	for _, f := range files {
		logger.Printf("removing plugin: %s\n", f)
		_ = os.Remove(f)
	}

	return nil
}

// == Helper Functions ===============================================

func errTypeToString(err error) string {
	if err == nil {
		return "<nil>"
	}

	return reflect.TypeOf(err).String()
}

// verboseEnabled allows us to see if the flag is set.  Unfortunately
// this is only available on *testing.T.Verbose() and not on *testing.M
func verboseEnabled() bool {
	for _, v := range os.Args {
		if v == "-test.v=true" {
			return true
		}
	}

	return false
}

// buildTestPlugin will build the plugin
func buildTestPlugins() error {

	paths, err := getTestPluginPaths()

	if err != nil {
		return fmt.Errorf("could not find matching plugins: %w", err)
	}

	for _, p := range paths {
		ctx, cancel := context.WithTimeout(context.Background(), buildPluginTimeout)
		defer cancel()

		baseName := strings.TrimSuffix(filepath.Base(p), ".go")
		dir := filepath.Dir(p)

		args := []string{
			"build",
			"-buildmode=plugin",
		}

		if raceFlagEnabled {
			args = append(args, "-race")
		}

		args = append(
			args,
			"-o", filepath.Join(dir, baseName+".so"),
			"./"+dir, // Must have the preceding "./" which a filepath.Join ends up removing
		)

		cmd := exec.CommandContext(ctx, "go", args...)

		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		logger.Printf("compiling plugin: %s", p)

		err := cmd.Run()

		if err != nil {
			return fmt.Errorf("error compiling plugin %s: %w", p, err)
		}

	}

	return nil

}

func getTestPluginPaths() ([]string, error) {
	pluginRegex := filepath.Join(testPluginDir, "*", "*plugin.go")
	return filepath.Glob(pluginRegex)
}

func getTestPluginSOPaths() ([]string, error) {
	pluginRegex := filepath.Join(testPluginDir, "*", "*.so")
	return filepath.Glob(pluginRegex)

}

// == Helper Structures ==============================================

// TODO: These may be able to be deprecated due to the new plugin.Plugin
// structure.
type newPluginererMock struct {
	plugin.NewPluginererMock
}

func (m *newPluginererMock) Name() string     { return "newPluginerMock" }
func (m *newPluginererMock) Version() string  { return "pluginVersion" }
func (m *newPluginererMock) CommitID() string { return "pluginCommitID" }
func (m *newPluginererMock) Config() string   { return "pluginConfig" }
func (m *newPluginererMock) SupportedTypes() bit.Mask {
	return plugin.TypePluginer
}

type newReceivererMock struct {
	receiver.NewReceivererMock
}

func (m *newReceivererMock) Name() string     { return "newReceivererMock" }
func (m *newReceivererMock) Version() string  { return "receiverVersion" }
func (m *newReceivererMock) CommitID() string { return "receiverCommitID" }
func (m *newReceivererMock) Config() string   { return "receiverConfig" }
func (m *newReceivererMock) SupportedTypes() bit.Mask {
	return plugin.TypeReceiver | plugin.TypePluginer
}

type newSendererMock struct {
	sender.NewSendererMock
}

func (m *newSendererMock) Name() string     { return "newSendererMock" }
func (m *newSendererMock) Version() string  { return "senderVersion" }
func (m *newSendererMock) CommitID() string { return "senderCommitID" }
func (m *newSendererMock) Config() string   { return "senderConfig" }
func (m *newSendererMock) SupportedTypes() bit.Mask {
	return plugin.TypeSender | plugin.TypePluginer
}

type newFiltererMock struct {
	filter.NewFiltererMock
}

func (m *newFiltererMock) Name() string     { return "newFiltererMock" }
func (m *newFiltererMock) Version() string  { return "filterVersion" }
func (m *newFiltererMock) CommitID() string { return "filterCommitID" }
func (m *newFiltererMock) Config() string   { return "filterConfig" }
func (m *newFiltererMock) SupportedTypes() bit.Mask {
	return plugin.TypeFilter | plugin.TypePluginer
}
