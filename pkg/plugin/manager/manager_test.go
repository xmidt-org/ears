// Licensed to Comcast Cable Communications Management, LLC under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Comcast Cable Communications Management, LLC licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package plugin_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xmidt-org/ears/internal/pkg/app/plugin"

	. "github.com/onsi/gomega"
)

const (
	buildPluginTimeout = 10 * time.Second

	testPluginDir  = "test_plugin"
	testPluginFile = "test_plugin.so"
)

var (
	testPluginPath = filepath.Join(testPluginDir, testPluginFile)
)

func TestMain(m *testing.M) {
	err := testSetup()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	code := m.Run()

	err = testShutdown()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1 | code)
	}

	os.Exit(code)
}

func TestLoadPlugin(t *testing.T) {
	testCases := []struct {
		name   string
		config plugin.Config
		err    error
	}{
		{
			name: "found",
			config: plugin.Config{
				Name: "found_plugin",
				Path: testPluginPath,
			},
		},
	}

	m := plugin.NewManager()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			p, err := m.LoadPlugin(tc.config)

			if tc.err == nil {
				a.Expect(err).To(BeNil())
				a.Expect(p).ToNot(BeNil())
			} else {
				a.Expect(err).To(Equal(tc.err))
			}
		})
	}

	fmt.Println(testCases)
}

func TestBuildTestPlugin(t *testing.T) {
	err := buildTestPlugin()
	if err != nil {
		t.Error(err)
	}
}

// === Test Lifecycle ===========================

func testSetup() error {
	return buildTestPlugin()
}

func testShutdown() error {
	_ = os.Remove(filepath.Join(testPluginDir, testPluginFile))
	return nil
}

// buildTestPlugin will build the plugin
func buildTestPlugin() error {

	ctx, cancel := context.WithTimeout(context.Background(), buildPluginTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"go", "build",
		"-buildmode=plugin",
		"-o", filepath.Join(testPluginDir, testPluginFile),
		"./"+testPluginDir, // Must have the preceeding "./" which a filepath.Join ends up removing
	)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()

}
