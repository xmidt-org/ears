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
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/plugin/manager"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"

	. "github.com/onsi/gomega"

	"github.com/sebdah/goldie/v2"
)

const (
	buildPluginTimeout = 10 * time.Second

	testPluginDir = "testplugins"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func TestMain(m *testing.M) {
	err := testSetup()
	if err != nil {
		logger.Fatal(err)
	}

	code := m.Run()

	err = testShutdown()
	if err != nil {
		logger.Println(err)
		os.Exit(1 | code)
	}

	os.Exit(code)
}

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "NilPluginError", err: &manager.NilPluginError{}},
		{name: "NotFoundError", err: &manager.NotFoundError{}},
		{name: "AlreadyRegisteredError", err: &manager.AlreadyRegisteredError{}},
		{
			name: "InvalidConfigError_Nil",
			err:  &manager.InvalidConfigError{},
		},
		{
			name: "InvalidConfigError_Err",
			err: &manager.InvalidConfigError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
		{name: "NewPluginerNotImplementedError", err: &manager.NewPluginerNotImplementedError{}},
		{
			name: "VariableLookupError_Nil",
			err:  &manager.VariableLookupError{},
		},
		{
			name: "VariableLookupError_Err",
			err: &manager.VariableLookupError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
		{
			name: "LoadError_Nil",
			err:  &manager.LoadError{},
		},
		{
			name: "LoadError_Err",
			err: &manager.LoadError{
				Err: fmt.Errorf("wrapped error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t)
			g.Assert(t, tc.name, []byte(fmt.Sprint(tc.err)))
			g.Assert(t, tc.name+"_unwrapped", []byte(fmt.Sprint(errors.Unwrap(tc.err))))

		})
	}

}

func TestNilPluginError(t *testing.T) {
	testCases := []struct {
		name string
		plug plugin.Pluginer
		err  error
	}{
		{name: "nil_plugin", err: &manager.NilPluginError{}},
	}

	m := manager.NewManager()

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
		{name: "registered", plug: &plugin.PluginerMock{}},
		{name: "registered", plug: &plugin.PluginerMock{}, err: &manager.AlreadyRegisteredError{}},
	}

	m := manager.NewManager()

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
		{name: "one", plug: &plugin.PluginerMock{}},
		{name: "two", plug: &plugin.PluginerMock{}},
		{name: "three", plug: &plugin.PluginerMock{}},
		{name: "four", plug: &plugin.PluginerMock{}},
	}

	m := manager.NewManager()
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
			m := manager.NewManager()

			// Plugins
			for i := 0; i < tc.numPlugins; i++ {
				err := m.RegisterPlugin(
					"plugins_"+strconv.Itoa(i),
					&plugin.PluginerMock{},
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

	m := manager.NewManager()
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
				}
			}

			if expectedErr == nil {
				a.Expect(err).To(BeNil())
				a.Expect(p).ToNot(BeNil())
			} else {
				a.Expect(p).To(BeNil())
				a.Expect(reflect.TypeOf(err)).To(Equal(reflect.TypeOf(expectedErr)))
			}

		})
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

// buildTestPlugin will build the plugin
func buildTestPlugins() error {

	ctx, cancel := context.WithTimeout(context.Background(), buildPluginTimeout)
	defer cancel()

	paths, err := getTestPluginPaths()

	if err != nil {
		return fmt.Errorf("could not find matching plugins: %w", err)
	}

	for _, p := range paths {
		baseName := strings.TrimSuffix(filepath.Base(p), ".go")
		dir := filepath.Dir(p)

		cmd := exec.CommandContext(
			ctx,
			"go", "build",
			"-buildmode=plugin",
			"-o", filepath.Join(dir, baseName+".so"),
			"./"+dir, // Must have the preceeding "./" which a filepath.Join ends up removing
		)

		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		logger.Printf("compiling plugin: %s", p)

		err := cmd.Run()

		if err != nil {
			return fmt.Errorf("error compiling plugin %s: %w", baseName, err)
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

type newReceivererMock struct {
	receiver.NewReceivererMock
}

func (m *newReceivererMock) Name() string    { return "newReceivererMock" }
func (m *newReceivererMock) Version() string { return "version" }
func (m *newReceivererMock) Config() string  { return "config" }

type newSendererMock struct {
	sender.NewSendererMock
}

func (m *newSendererMock) Name() string    { return "newSendererMock" }
func (m *newSendererMock) Version() string { return "version" }
func (m *newSendererMock) Config() string  { return "config" }

type newFiltererMock struct {
	filter.NewFiltererMock
}

func (m *newFiltererMock) Name() string    { return "newFiltererMock" }
func (m *newFiltererMock) Version() string { return "version" }
func (m *newFiltererMock) Config() string  { return "config" }
