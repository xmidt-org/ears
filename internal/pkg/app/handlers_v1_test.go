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

package app

import (
	"encoding/json"
	goldie "github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/plugin"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/plugin/manager"
	"github.com/xmidt-org/ears/pkg/plugins/block"
	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xmidt-org/ears/pkg/plugins/match"
	"github.com/xmidt-org/ears/pkg/plugins/pass"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestVersionHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/version", nil)
	api, err := NewAPIManager(&DefaultRoutingTableManager{})
	if err != nil {
		t.Fatalf("Fail to instaniate api manager: %s\n", err.Error())
	}
	api.versionHandler(w, r)
	g := goldie.New(t)
	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "version", data)
}

func TestAddRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	// rename to .json
	simpleRouteReader, err := os.Open("testdata/route.golden")
	if err != nil {
		t.Fatalf("cannot read testdata/route.golden")
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	inMemStorageMgr := db.NewInMemoryRouteStorer(nil)
	mgr, err := manager.New()
	if err != nil {
		t.Fatalf("cannot create plugin manager: %s\n", err.Error())
	}
	pluginMgr, err := plugin.NewManager(plugin.WithPluginManager(mgr))
	// register plugins
	toArr := func(a ...interface{}) []interface{} { return a }
	defaultPlugins := []struct {
		name   string
		plugin pkgplugin.Pluginer
	}{
		{
			name:   "debug",
			plugin: toArr(debug.NewPluginVersion("debug", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "match",
			plugin: toArr(match.NewPluginVersion("match", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "pass",
			plugin: toArr(pass.NewPluginVersion("pass", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "block",
			plugin: toArr(block.NewPluginVersion("block", "", ""))[0].(pkgplugin.Pluginer),
		},
	}
	for _, plug := range defaultPlugins {
		err = mgr.RegisterPlugin(plug.name, plug.plugin)
		if err != nil {
			t.Fatalf("could register %s plugin: %s", plug.name, err.Error())
		}
	}
	//
	routingMgr := NewRoutingTableManager(pluginMgr, inMemStorageMgr)
	api, err := NewAPIManager(routingMgr)
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	api.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addroute", data)
}
