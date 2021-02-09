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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	goldie "github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/plugin"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/plugin/manager"
	"github.com/xmidt-org/ears/pkg/plugins/block"
	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xmidt-org/ears/pkg/plugins/match"
	"github.com/xmidt-org/ears/pkg/plugins/pass"
	"github.com/xmidt-org/ears/pkg/plugins/split"
	"github.com/xmidt-org/ears/pkg/route"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func setupRestApi() (*APIManager, RoutingTableManager, plugin.Manager, route.RouteStorer, error) {
	inMemStorageMgr := db.NewInMemoryRouteStorer(nil)
	mgr, err := manager.New()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	pluginMgr, err := plugin.NewManager(plugin.WithPluginManager(mgr))
	if err != nil {
		return nil, nil, nil, nil, err
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
		{
			name:   "split",
			plugin: toArr(split.NewPluginVersion("split", "", ""))[0].(pkgplugin.Pluginer),
		},
	}
	for _, plug := range defaultPlugins {
		err = mgr.RegisterPlugin(plug.name, plug.plugin)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}
	routingMgr := NewRoutingTableManager(pluginMgr, inMemStorageMgr)
	apiMgr, err := NewAPIManager(routingMgr)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return apiMgr, routingMgr, pluginMgr, inMemStorageMgr, nil

}

func checkEventsSent(waitSeconds int, routeFileName string, pluginMgr plugin.Manager, expectedNumberOfEvents int, eventFileName string, eventIndex int) error {
	time.Sleep(time.Duration(waitSeconds) * time.Second)
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	buf, err := ioutil.ReadFile(routeFileName)
	if err != nil {
		return err
	}
	var rt route.Config
	err = json.Unmarshal(buf, &rt)
	if err != nil {
		return err
	}
	sdr, err := pluginMgr.RegisterSender(ctx, rt.Sender.Plugin, rt.Sender.Name, stringify(rt.Sender.Config))
	if err != nil {
		return err
	}
	debugSender, ok := sdr.Unwrap().(*debug.Sender)
	if !ok {
		return errors.New("bad type assertion debug sender")
	}
	if debugSender.Count() != expectedNumberOfEvents {
		return errors.New(fmt.Sprintf("unexpected number of events in sender %d (%d)", debugSender.Count(), expectedNumberOfEvents))
	}
	// spot check event payload if desired
	if eventFileName != "" && eventIndex >= 0 {
		events := debugSender.History()
		if events == nil || len(events) == 0 {
			return errors.New("no debug events collected")
		}
		if eventIndex >= len(events) {
			return errors.New(fmt.Sprintf("event index %d out of range (%d)", eventIndex, len(events)))
		}
		buf1, err := json.Marshal(events[eventIndex].Payload())
		if err != nil {
			return err
		}
		buf2, err := ioutil.ReadFile(eventFileName)
		if err != nil {
			return err
		}
		var gevt interface{}
		err = json.Unmarshal(buf2, &gevt)
		if err != nil {
			return err
		}
		buf2, err = json.Marshal(gevt)
		if err != nil {
			return err
		}
		if string(buf1) != string(buf2) {
			return errors.New(fmt.Sprintf("event payload mismatch:\n%s\n%s\n", string(buf1), string(buf2)))
		}
	}
	return nil
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

// single route tests

func TestRestPostSimpleRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, pluginMgr, _, err := setupRestApi()
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
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

func TestRestPutSimpleRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	name := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(name)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPut, "/ears/v1/routes/r123", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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

func TestRestPostFilterMatchAllowRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterMatchAllowRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, pluginMgr, _, err := setupRestApi()
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
	g.AssertJson(t, "addfiltermatchallowroute", data)
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

func TestRestPostFilterMatchDenyRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterMatchDenyRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, pluginMgr, _, err := setupRestApi()
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
	g.AssertJson(t, "addfiltermatchdenyroute", data)
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 0, "", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

func TestRestPostFilterChainMatchRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterChainMatchRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, pluginMgr, _, err := setupRestApi()
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
	g.AssertJson(t, "addfilterchainmatchroute", data)
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 5, "testdata/event2.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

func TestRestPostFilterSplitRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterSplitRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, pluginMgr, _, err := setupRestApi()
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
	g.AssertJson(t, "addsimplefiltersplitroute", data)
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 10, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

func TestRestPostFilterDeepSplitRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterDeepSplitRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, pluginMgr, _, err := setupRestApi()
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
	g.AssertJson(t, "addsimplefilterdeepsplitroute", data)
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 10, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

// multi route tests

func TestRestMultiRouteAABB(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteAA.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	routeFileName2 := "testdata/simpleRouteBB.json"
	simpleRouteReader2, err := os.Open(routeFileName2)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	api, _, pluginMgr, _, err := setupRestApi()
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader2)
	api.muxRouter.ServeHTTP(w, r)
	/*g := goldie.New(t)
	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addroute", data)*/
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	err = checkEventsSent(0, routeFileName2, pluginMgr, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

func TestRestMultiRouteAABBAB(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteAA.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	routeFileName2 := "testdata/simpleRouteBB.json"
	simpleRouteReader2, err := os.Open(routeFileName2)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	routeFileName3 := "testdata/simpleRouteAB.json"
	simpleRouteReader3, err := os.Open(routeFileName3)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	api, _, pluginMgr, _, err := setupRestApi()
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader2)
	api.muxRouter.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader3)
	api.muxRouter.ServeHTTP(w, r)
	/*g := goldie.New(t)
	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addroute", data)*/
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	err = checkEventsSent(0, routeFileName2, pluginMgr, 10, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

func TestRestMultiRouteAAAB(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteAA.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	routeFileName2 := "testdata/simpleRouteAB.json"
	simpleRouteReader2, err := os.Open(routeFileName2)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	api, _, pluginMgr, _, err := setupRestApi()
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader2)
	api.muxRouter.ServeHTTP(w, r)
	/*g := goldie.New(t)
	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addroute", data)*/
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	err = checkEventsSent(0, routeFileName2, pluginMgr, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

func TestRestMultiRouteBBAB(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteBB.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	routeFileName2 := "testdata/simpleRouteAB.json"
	simpleRouteReader2, err := os.Open(routeFileName2)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	routeFileName3 := "testdata/simpleRouteAA.json"
	api, _, pluginMgr, _, err := setupRestApi()
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader2)
	api.muxRouter.ServeHTTP(w, r)
	/*g := goldie.New(t)
	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addroute", data)*/
	// check number of events received by output plugin
	err = checkEventsSent(1, routeFileName, pluginMgr, 10, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	err = checkEventsSent(0, routeFileName3, pluginMgr, 0, "", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
}

// various api tests

func TestRestGetRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	routeFileName := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	api, _, _, _, err := setupRestApi()
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
	routeFileName := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	/*simpleFilterRouteReader, err := os.Open("testdata/simpleFilterMatchAllowRoute.json")
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}*/
	api, _, _, _, err := setupRestApi()
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api.muxRouter.ServeHTTP(w, r)
	/*w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleFilterRouteReader)
	api.muxRouter.ServeHTTP(w, r)*/
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
	routeFileName := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	routeFileName2 := "testdata/simpleFilterMatchAllowRoute.json"
	simpleFilterRouteReader, err := os.Open(routeFileName2)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	api, _, _, _, err := setupRestApi()
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

// tests for various error conditions

func TestRestRouteHandlerIdMismatch(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPut, "/ears/v1/routes/badid", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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
	api, _, _, _, err := setupRestApi()
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

func TestRestPostRouteHandlerBadName(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteBadName.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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
	g.AssertJson(t, "addroutebadname", data)
}

func TestRestPostRouteHandlerBadPluginName(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteBadPluginName.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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
	g.AssertJson(t, "addroutebadpluginname", data)
}

func TestRestPostRouteHandlerNoApp(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteNoApp.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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
	g.AssertJson(t, "addroutenoapp", data)
}

func TestRestPostRouteHandlerNoOrg(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteNoOrg.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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
	g.AssertJson(t, "addroutenoorg", data)
}

func TestRestPostRouteHandlerNoSender(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteNoSender.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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
	g.AssertJson(t, "addroutenosender", data)
}

func TestRestPostRouteHandlerNoReceiver(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteNoReceiver.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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
	g.AssertJson(t, "addroutenoreceiver", data)
}

func TestRestPostRouteHandlerNoUser(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteNoUser.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	r := httptest.NewRequest(http.MethodPost, "/ears/v1/routes", simpleRouteReader)
	api, _, _, _, err := setupRestApi()
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
	g.AssertJson(t, "addroutenouser", data)
}
