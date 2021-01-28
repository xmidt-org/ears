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

func setupRestApi() (*APIManager, error) {
	inMemStorageMgr := db.NewInMemoryRouteStorer(nil)
	mgr, err := manager.New()
	if err != nil {
		return nil, err
	}
	pluginMgr, err := plugin.NewManager(plugin.WithPluginManager(mgr))
	if err != nil {
		return nil, err
	}
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
			return nil, err
		}
	}
	routingMgr := NewRoutingTableManager(pluginMgr, inMemStorageMgr)
	return NewAPIManager(routingMgr)

}

func TestRestVersionHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/version", nil)
	api, err := NewAPIManager(&DefaultRoutingTableManager{})
	if err != nil {
		t.Fatalf("Fail to setup api manager: %s\n", err.Error())
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

func TestRestPostRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	simpleRouteReader, err := os.Open("testdata/simpleRoute.json")
	if err != nil {
		t.Fatalf("cannot read route.json")
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, err := setupRestApi()
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

func TestRestPutRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	simpleRouteReader, err := os.Open("testdata/simpleRoute.json")
	if err != nil {
		t.Fatalf("cannot read route.json")
	}
	r := httptest.NewRequest(http.MethodPut, "/ears/v1/routes/r123", simpleRouteReader)
	api, err := setupRestApi()
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

func TestRestRouteHandlerIdMismatch(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	simpleRouteReader, err := os.Open("testdata/simpleRoute.json")
	if err != nil {
		t.Fatalf("cannot read route.json")
	}
	r := httptest.NewRequest(http.MethodPut, "/ears/v1/routes/badid", simpleRouteReader)
	api, err := setupRestApi()
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
	g.AssertJson(t, "addrouteidmismatch", data)
}

func TestRestMissingRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ears/v1/routes/fakeid", nil)
	api, err := setupRestApi()
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
	g.AssertJson(t, "missingroute", data)
}

func TestRestGetRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	simpleRouteReader, err := os.Open("testdata/simpleRoute.json")
	if err != nil {
		t.Fatalf("cannot read route.json")
	}
	api, err := setupRestApi()
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/ears/v1/routes/r123", nil)
	api.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	item := data["item"].(map[string]interface{})
	delete(item, "created")
	delete(item, "modified")
	g.AssertJson(t, "getroute", data)
}

func TestRestGetMultipleRoutesHandler(t *testing.T) {
	Version = "v1.0.2"
	simpleRouteReader, err := os.Open("testdata/simpleRoute.json")
	if err != nil {
		t.Fatalf("cannot read route.json")
	}
	simpleFilterRouteReader, err := os.Open("testdata/simpleFilterRoute.json")
	if err != nil {
		t.Fatalf("cannot read route.json")
	}
	api, err := setupRestApi()
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleFilterRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/ears/v1/routes", nil)
	api.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	items := data["items"].([]interface{})
	for _, item := range items {
		delete(item.(map[string]interface{}), "created")
		delete(item.(map[string]interface{}), "modified")
	}
	g.AssertJson(t, "getmultipleroutes", data)
}

func TestRestDeleteRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	simpleRouteReader, err := os.Open("testdata/simpleRoute.json")
	if err != nil {
		t.Fatalf("cannot read route.json")
	}
	simpleFilterRouteReader, err := os.Open("testdata/simpleFilterRoute.json")
	if err != nil {
		t.Fatalf("cannot read route.json")
	}
	api, err := setupRestApi()
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/ears/v1/routes/r123", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleFilterRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1/routes/r123", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/ears/v1/routes", nil)
	api.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	items := data["items"].([]interface{})
	for _, item := range items {
		delete(item.(map[string]interface{}), "created")
		delete(item.(map[string]interface{}), "modified")
	}
	g.AssertJson(t, "deleteroute", data)
}
