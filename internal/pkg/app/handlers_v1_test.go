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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	goldie "github.com/sebdah/goldie/v2"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/db/bolt"
	"github.com/xmidt-org/ears/internal/pkg/db/dynamo"
	"github.com/xmidt-org/ears/internal/pkg/db/redis"
	"github.com/xmidt-org/ears/internal/pkg/plugin"
	"github.com/xmidt-org/ears/internal/pkg/tablemgr"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/plugin/manager"
	"github.com/xmidt-org/ears/pkg/plugins/block"
	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xmidt-org/ears/pkg/plugins/match"
	"github.com/xmidt-org/ears/pkg/plugins/pass"
	"github.com/xmidt-org/ears/pkg/plugins/split"
	"github.com/xmidt-org/ears/pkg/plugins/transform"
	"github.com/xmidt-org/ears/pkg/plugins/unwrap"
	"github.com/xmidt-org/ears/pkg/route"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type (
	RouteTestTable struct {
		Table                   map[string]*RouteTest `json:"table,omitempty"`
		SharePluginsAcrossTests bool                  `json:"sharePluginsAcrossTests,omitempty"` // name the effect no the cause
		TestToRunAllIfBlank     string                `json:"testToRunAllIfBlank,omitempty"`     // the self documenting variable name
		NumInstances            int                   `json:"numInstances,omitempty"`            // number of ears instances (should be between 1 and, say, 5)
		StorageType             string                `json:"storageType,omitempty"`             // if blank use whatever is specified in ears.yaml
	}
	RouteTest struct {
		SequenceNumber int              `json:"seq,omitempty"`
		RouteFiles     []string         `json:"routeFiles,omitempty"`
		WaitMs         int              `json:"waitMs,omitempty"`
		Events         []EventCheckTest `json:"events,omitempty"`
	}
	EventCheckTest struct {
		SenderRouteFile          string `json:"senderRouteFiles,omitempty"`
		ExpectedEventCount       int    `json:"expectedEventCount,omitempty"`
		ExpectedEventIndex       int    `json:"expectedEventIndex,omitempty"`
		ExpectedEventPayloadFile string `json:"expectedEventPayloadFile,omitempty"`
	}
	EarsRuntime struct {
		config              config.Config
		apiManager          *APIManager
		pluginManger        plugin.Manager
		storageLayer        route.RouteStorer
		routingTableManager tablemgr.RoutingTableManager
	}
)

const tenantPath = "/orgs/comcast/applications/xfi"

// should in memory storer also be implemented as singleton?

var cachedInMemoryStorageLayer route.RouteStorer

func stringify(data interface{}) string {
	if data == nil {
		return ""
	}
	buf, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(buf)
}

func prefixRouteConfig(routeConfig *route.Config, prefix string) {
	routeConfig.Name = prefix + routeConfig.Name
	if routeConfig.Sender.Name != "" {
		routeConfig.Sender.Name = prefix + routeConfig.Sender.Name
	}
	if routeConfig.Receiver.Name != "" {
		routeConfig.Receiver.Name = prefix + routeConfig.Receiver.Name
	}
	if routeConfig.FilterChain != nil {
		for _, f := range routeConfig.FilterChain {
			if f.Name != "" {
				f.Name = prefix + f.Name
			}
		}
	}
}

func TestRouteTable(t *testing.T) {
	// global test settings
	Version = "v1.0.2"
	testTableName := "table"
	// load test table
	testTableFileName := "testdata/" + testTableName + ".json"
	buf, err := ioutil.ReadFile(testTableFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	var table RouteTestTable
	err = json.Unmarshal(buf, &table)
	if err != nil {
		t.Fatalf("cannot parse test table: %s", err.Error())
	}
	if table.NumInstances < 1 || table.NumInstances > 5 {
		t.Fatalf("number of ears instances must be between 1 and 5")
	}
	// setup ears runtime
	config, err := getConfig()
	if err != nil {
		t.Fatalf("cannot get config: %s", err.Error())
	}
	storageMgr, err := getStorageLayer(config, table.StorageType)
	if err != nil {
		t.Fatalf("cannot get stroage manager: %s", err.Error())
	}
	runtime, err := setupRestApi(config, storageMgr)
	if err != nil {
		t.Fatalf("cannot create ears runtime: %s\n", err.Error())
	}
	runtime.routingTableManager.StartListeningForSyncRequests("")
	// add passive ears instances if any
	passiveRuntimes := make([]*EarsRuntime, 0)
	if table.NumInstances < 2 {
		t.Logf("no passive ears runtime configured")
	}
	for i := 1; i < table.NumInstances; i++ {
		rt, err := setupRestApi(config, storageMgr)
		if err != nil {
			t.Fatalf("cannot create passive ears runtime: %s\n", err.Error())
		}
		// should review this
		if getTableSyncerType(config, "") != "inmemory" {
			rt.routingTableManager.StartListeningForSyncRequests("")
		} else {
			runtime.routingTableManager.RegisterLocalTableSyncer(rt.routingTableManager)
		}
		t.Logf("started passive ears runtime %d", i)
		passiveRuntimes = append(passiveRuntimes, rt)
	}
	// run tests
	cnt := 0
	for currentTestName, currentTest := range table.Table {
		if table.TestToRunAllIfBlank != "" && table.TestToRunAllIfBlank != currentTestName {
			continue
		}
		cnt++
		t.Run(currentTestName, func(t *testing.T) {
			testPrefix := ""
			if !table.SharePluginsAcrossTests {
				testPrefix = "tbltst" + currentTestName
			}
			t.Logf("SCENARIO: %s [id=%d] [cnt=%d]", currentTestName, currentTest.SequenceNumber, cnt)
			// setup routes
			routeIds := make([]string, 0)
			for _, routeFileName := range currentTest.RouteFiles {
				// read and parse route
				buf, err := ioutil.ReadFile("testdata/" + routeFileName + ".json")
				if err != nil {
					t.Fatalf("%s test: cannot read route file: %s", currentTestName, err.Error())
				}
				// scope route by prefixing all names (confirm with Trevor what unregister is meant to do)
				var routeConfig route.Config
				err = json.Unmarshal(buf, &routeConfig)
				if err != nil {
					t.Fatalf("%s test: cannot parse route: %s", currentTestName, err.Error())
				}
				prefixRouteConfig(&routeConfig, testPrefix)
				buf, err = json.MarshalIndent(routeConfig, "", "\t")
				if err != nil {
					t.Fatalf("%s test: cannot serialize route: %s", currentTestName, err.Error())
				}
				// add route
				r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", bytes.NewReader(buf))
				w := httptest.NewRecorder()
				runtime.apiManager.muxRouter.ServeHTTP(w, r)
				g := goldie.New(t)
				var data Response
				err = json.Unmarshal(w.Body.Bytes(), &data)
				if err != nil {
					t.Fatalf("%s test: cannot unmarshal response: %s %s", currentTestName, err.Error(), string(w.Body.Bytes()))
				}
				g.AssertJson(t, "tbl_"+currentTestName+"_"+routeConfig.Name, data)
				// collect route ID
				if data.Item == nil {
					t.Fatalf("%s test: no item in response", currentTestName)
				}
				buf, err = json.Marshal(data.Item)
				if err != nil {
					t.Fatalf("%s test: %s", currentTestName, err.Error())
				}
				var rt route.Config
				err = json.Unmarshal(buf, &rt)
				if err != nil {
					fmt.Printf("%v\n", data.Item)
					t.Fatalf("%s test: item is not a route: %s", currentTestName, err.Error())
				}
				if rt.Id == "" {
					t.Fatalf("%s test: route has blank ID", currentTestName)
				}
				routeIds = append(routeIds, rt.Id)
				t.Logf("added route with tid: %s, id: %s", rt.TenantId.String(), rt.Id)
			}
			// sleep
			time.Sleep(time.Duration(currentTest.WaitMs) * time.Millisecond)
			// check number of routes in persistence layer
			err = checkNumRoutes(runtime.apiManager, currentTestName, len(routeIds))
			if err != nil {
				t.Fatalf("%s test: route count issue: %s", currentTestName, err.Error())
			}
			for _, rt := range passiveRuntimes {
				err = checkNumRoutes(rt.apiManager, currentTestName, len(routeIds))
				if err != nil {
					t.Fatalf("%s test: synchronized route count issue: %s", currentTestName, err.Error())
				}
			}
			// check number of registered / running routes
			registeredRoutes, _ := runtime.routingTableManager.GetAllRegisteredRoutes()
			if len(registeredRoutes) != len(routeIds) {
				t.Fatalf("%s test: registered route count mismatch: %d (%d)", currentTestName, len(registeredRoutes), len(routeIds))
			}
			for _, rt := range passiveRuntimes {
				registeredRoutes, _ = rt.routingTableManager.GetAllRegisteredRoutes()
				if len(registeredRoutes) != len(routeIds) {
					t.Fatalf("%s test: synchronized registered route count mismatch: %d (%d)", currentTestName, len(registeredRoutes), len(routeIds))
				}
			}
			// sleep
			time.Sleep(time.Duration(currentTest.WaitMs) * time.Millisecond)
			// check number of events and payloads if desired
			for _, eventData := range currentTest.Events {
				if eventData.SenderRouteFile == "" {
					continue
				}
				routeFileName := "testdata/" + eventData.SenderRouteFile + ".json"
				eventFileName := ""
				if eventData.ExpectedEventPayloadFile != "" {
					eventFileName = "testdata/" + eventData.ExpectedEventPayloadFile + ".json"
				}
				err = checkEventsSent(routeFileName, testPrefix, runtime.pluginManger, eventData.ExpectedEventCount, eventFileName, eventData.ExpectedEventIndex)
				if err != nil {
					t.Fatalf("%s test: check events sent error: %s", currentTestName, err.Error())
				}
			}
			// delete all routes
			for _, rtId := range routeIds {
				r := httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rtId, nil)
				w := httptest.NewRecorder()
				runtime.apiManager.muxRouter.ServeHTTP(w, r)
				t.Logf("deleted route with id: %s", rtId)
			}
			// sleep
			time.Sleep(time.Duration(currentTest.WaitMs) * time.Millisecond)
			// check number of routes in persistence layer
			err = checkNumRoutes(runtime.apiManager, currentTestName, 0)
			if err != nil {
				t.Fatalf("%s test: zero route count issue: %s", currentTestName, err.Error())
			}
			for _, rt := range passiveRuntimes {
				err = checkNumRoutes(rt.apiManager, currentTestName, 0)
				if err != nil {
					t.Fatalf("%s test: synchronized route count issue: %s", currentTestName, err.Error())
				}
			}
			// check number of registered / running routes
			registeredRoutes, _ = runtime.routingTableManager.GetAllRegisteredRoutes()
			if len(registeredRoutes) != 0 {
				t.Fatalf("%s test: registered route count mismatch: %d (%d)", currentTestName, len(registeredRoutes), len(routeIds))
			}
			for _, rt := range passiveRuntimes {
				registeredRoutes, _ = rt.routingTableManager.GetAllRegisteredRoutes()
				if len(registeredRoutes) != 0 {
					t.Fatalf("%s test: synchronized registered route count mismatch: %d (%d)", currentTestName, len(registeredRoutes), len(routeIds))
				}
			}
			// sleep
			time.Sleep(time.Duration(currentTest.WaitMs) * time.Millisecond)
		})
	}
	// tear down ears runtime
	runtime.routingTableManager.StopListeningForSyncRequests("")
	for _, rt := range passiveRuntimes {
		rt.routingTableManager.StopListeningForSyncRequests("")
	}
}

func checkNumRoutes(api *APIManager, currentTestName string, numExpected int) error {
	r := httptest.NewRequest(http.MethodGet, "/ears/v1"+tenantPath+"/routes", nil)
	w := httptest.NewRecorder()
	api.muxRouter.ServeHTTP(w, r)
	var data Response
	var err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		return errors.New(fmt.Sprintf("%s test: cannot unmarshal response: %s %s", currentTestName, err.Error(), string(w.Body.Bytes())))
	}
	if data.Items == nil {
		return errors.New(fmt.Sprintf("%s test: no items found", currentTestName))
	}
	itemsArray, ok := data.Items.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("%s test: items not an array", currentTestName))
	}
	if len(itemsArray) != numExpected {
		return errors.New(fmt.Sprintf("%s test: unexpected number of items %d (%d)", currentTestName, len(itemsArray), numExpected))
	}
	return nil
}

func setupSimpleApi(t *testing.T, storageType string) *EarsRuntime {
	config, err := getConfig()
	if err != nil {
		t.Fatalf("cannot get config: %s", err.Error())
	}
	storageMgr, err := getStorageLayer(config, storageType)
	if err != nil {
		t.Fatalf("cannot get stroage manager: %s", err.Error())
	}
	runtime, err := setupRestApi(config, storageMgr)
	if err != nil {
		t.Fatalf("cannot create api manager: %s\n", err.Error())
	}
	return runtime
}

func getConfig() (config.Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("ears")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	config := viper.GetViper()
	return config, nil
}

// if storageType is blank choose storag elayer specified in ears.yaml
func getStorageLayer(config config.Config, storageType string) (route.RouteStorer, error) {
	if storageType == "" {
		storageType = config.GetString("ears.storage.type")
	}
	var storageMgr route.RouteStorer
	var err error
	switch storageType {
	case "inmemory":
		if cachedInMemoryStorageLayer != nil {
			storageMgr = cachedInMemoryStorageLayer
		} else {
			storageMgr = db.NewInMemoryRouteStorer(config)
			cachedInMemoryStorageLayer = storageMgr
		}
	case "dynamodb":
		storageMgr, err = dynamo.NewDynamoDbStorer(config)
		if err != nil {
			return nil, err
		}
	case "boltdb":
		storageMgr, err = bolt.NewBoltDbStorer(config)
		if err != nil {
			return nil, err
		}
	case "redis":
		storageMgr, err = redis.NewRedisDbStorer(config, &log.Logger)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported storage type '" + storageType + "'")
	}
	return storageMgr, nil
}

// if storageType is blank choose storag elayer specified in ears.yaml
func getTableSyncer(config config.Config, syncType string) (tablemgr.RoutingTableDeltaSyncer, error) {
	if syncType == "" {
		syncType = config.GetString("ears.synchronization.type")
	}
	var syncer tablemgr.RoutingTableDeltaSyncer
	switch syncType {
	case "inmemory":
		syncer = tablemgr.NewInMemoryDeltaSyncer(&log.Logger, config)
	case "redis":
		syncer = tablemgr.NewRedisDeltaSyncer(&log.Logger, config)
	default:
		return nil, errors.New("unsupported syncer type '" + syncType + "'")
	}
	return syncer, nil
}

func getTableSyncerType(config config.Config, syncType string) string {
	if syncType == "" {
		syncType = config.GetString("ears.synchronization.type")
	}
	return syncType
}

func setupRestApi(config config.Config, storageMgr route.RouteStorer) (*EarsRuntime, error) {
	mgr, err := manager.New()
	if err != nil {
		return &EarsRuntime{config, nil, nil, storageMgr, nil}, err
	}
	pluginMgr, err := plugin.NewManager(plugin.WithPluginManager(mgr), plugin.WithLogger(&log.Logger))
	if err != nil {
		return &EarsRuntime{config, nil, nil, storageMgr, nil}, err
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
		{
			name:   "unwrap",
			plugin: toArr(unwrap.NewPluginVersion("unwrap", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "transform",
			plugin: toArr(transform.NewPluginVersion("transform", "", ""))[0].(pkgplugin.Pluginer),
		},
	}
	for _, plug := range defaultPlugins {
		err = mgr.RegisterPlugin(plug.name, plug.plugin)
		if err != nil {
			return &EarsRuntime{config, nil, nil, storageMgr, nil}, err
		}
	}
	tableSyncer, err := getTableSyncer(config, "")
	if err != nil {
		return &EarsRuntime{config, nil, nil, storageMgr, nil}, err
	}
	routingMgr := tablemgr.NewRoutingTableManager(pluginMgr, storageMgr, tableSyncer, &log.Logger, config)
	apiMgr, err := NewAPIManager(routingMgr)
	if err != nil {
		return &EarsRuntime{config, nil, nil, storageMgr, nil}, err
	}
	return &EarsRuntime{config, apiMgr, pluginMgr, storageMgr, routingMgr}, nil
}

func resetDebugSender(routeFileName string, pluginMgr plugin.Manager) error {
	//zerolog.SetGlobalLevel(zerolog.ErrorLevel)
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
	debugSender.Reset()
	if debugSender.Count() != 0 {
		return errors.New(fmt.Sprintf("unexpected number of events in sender after reset %d (%d)", debugSender.Count(), 0))
	}
	err = pluginMgr.UnregisterSender(ctx, sdr)
	if err != nil {
		return err
	}
	return nil
}

func checkEventsSent(routeFileName string, testPrefix string, pluginMgr plugin.Manager, expectedNumberOfEvents int, eventFileName string, eventIndex int) error {
	//zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	buf, err := ioutil.ReadFile(routeFileName)
	if err != nil {
		return err
	}
	var routeConfig route.Config
	err = json.Unmarshal(buf, &routeConfig)
	if err != nil {
		return err
	}
	prefixRouteConfig(&routeConfig, testPrefix)
	sdr, err := pluginMgr.RegisterSender(ctx, routeConfig.Sender.Plugin, routeConfig.Sender.Name, stringify(routeConfig.Sender.Config))
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
	err = pluginMgr.UnregisterSender(ctx, sdr)
	if err != nil {
		return err
	}
	return nil
}

func TestRestVersionHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/version", nil)
	api, err := NewAPIManager(&tablemgr.DefaultRoutingTableManager{})
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
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data Response
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addpostroute", data)
	// check number of events received by output plugin
	time.Sleep(time.Duration(100) * time.Millisecond)
	err = checkEventsSent(routeFileName, "", runtime.pluginManger, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	// collect route ID
	if data.Item == nil {
		t.Fatalf("no item in response")
	}
	buf, err := json.Marshal(data.Item)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	var rt route.Config
	err = json.Unmarshal(buf, &rt)
	if err != nil {
		fmt.Printf("%v\n", data.Item)
		t.Fatalf("item is not a route: %s", err.Error())
	}
	if rt.Id == "" {
		t.Fatalf("route has blank ID")
	}
	// delete route
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rt.Id, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rt.Id)
}

func TestRestPutSimpleRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	name := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(name)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPut, "/ears/v1"+tenantPath+"/routes/r100", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data Response
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addputroute", data)
	// collect route ID
	if data.Item == nil {
		t.Fatalf("no item in response")
	}
	buf, err := json.Marshal(data.Item)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	var rt route.Config
	err = json.Unmarshal(buf, &rt)
	if err != nil {
		fmt.Printf("%v\n", data.Item)
		t.Fatalf("item is not a route: %s", err.Error())
	}
	if rt.Id == "" {
		t.Fatalf("route has blank ID")
	}
	// delete route
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rt.Id, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rt.Id)
}

func TestRestPostFilterMatchAllowRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterMatchAllowRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data Response
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addfiltermatchallowroute", data)
	// check number of events received by output plugin
	time.Sleep(time.Duration(100) * time.Millisecond)
	err = checkEventsSent(routeFileName, "", runtime.pluginManger, 5, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	// collect route ID
	if data.Item == nil {
		t.Fatalf("no item in response")
	}
	buf, err := json.Marshal(data.Item)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	var rt route.Config
	err = json.Unmarshal(buf, &rt)
	if err != nil {
		fmt.Printf("%v\n", data.Item)
		t.Fatalf("item is not a route: %s", err.Error())
	}
	if rt.Id == "" {
		t.Fatalf("route has blank ID")
	}
	// delete route
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rt.Id, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rt.Id)
}

func TestRestPostFilterMatchDenyRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterMatchDenyRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data Response
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addfiltermatchdenyroute", data)
	// check number of events received by output plugin
	time.Sleep(time.Duration(100) * time.Millisecond)
	err = checkEventsSent(routeFileName, "", runtime.pluginManger, 0, "", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	// collect route ID
	if data.Item == nil {
		t.Fatalf("no item in response")
	}
	buf, err := json.Marshal(data.Item)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	var rt route.Config
	err = json.Unmarshal(buf, &rt)
	if err != nil {
		fmt.Printf("%v\n", data.Item)
		t.Fatalf("item is not a route: %s", err.Error())
	}
	if rt.Id == "" {
		t.Fatalf("route has blank ID")
	}
	// delete route
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rt.Id, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rt.Id)
}

func TestRestPostFilterChainMatchRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterChainMatchRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data Response
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addfilterchainmatchroute", data)
	// check number of events received by output plugin
	time.Sleep(time.Duration(100) * time.Millisecond)
	err = checkEventsSent(routeFileName, "", runtime.pluginManger, 5, "testdata/event2.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	// collect route ID
	if data.Item == nil {
		t.Fatalf("no item in response")
	}
	buf, err := json.Marshal(data.Item)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	var rt route.Config
	err = json.Unmarshal(buf, &rt)
	if err != nil {
		fmt.Printf("%v\n", data.Item)
		t.Fatalf("item is not a route: %s", err.Error())
	}
	if rt.Id == "" {
		t.Fatalf("route has blank ID")
	}
	// delete route
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rt.Id, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rt.Id)
}

func TestRestPostFilterSplitRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterSplitRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data Response
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addsimplefiltersplitroute", data)
	// check number of events received by output plugin
	time.Sleep(time.Duration(100) * time.Millisecond)
	err = checkEventsSent(routeFileName, "", runtime.pluginManger, 10, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	// collect route ID
	if data.Item == nil {
		t.Fatalf("no item in response")
	}
	buf, err := json.Marshal(data.Item)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	var rt route.Config
	err = json.Unmarshal(buf, &rt)
	if err != nil {
		fmt.Printf("%v\n", data.Item)
		t.Fatalf("item is not a route: %s", err.Error())
	}
	if rt.Id == "" {
		t.Fatalf("route has blank ID")
	}
	// delete route
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rt.Id, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rt.Id)
}

func TestRestPostFilterDeepSplitRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleFilterDeepSplitRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data Response
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addsimplefilterdeepsplitroute", data)
	// check number of events received by output plugin
	time.Sleep(time.Duration(100) * time.Millisecond)
	err = checkEventsSent(routeFileName, "", runtime.pluginManger, 10, "testdata/event1.json", 0)
	if err != nil {
		t.Fatalf("check events sent error: %s", err.Error())
	}
	// collect route ID
	if data.Item == nil {
		t.Fatalf("no item in response")
	}
	buf, err := json.Marshal(data.Item)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	var rt route.Config
	err = json.Unmarshal(buf, &rt)
	if err != nil {
		fmt.Printf("%v\n", data.Item)
		t.Fatalf("item is not a route: %s", err.Error())
	}
	if rt.Id == "" {
		t.Fatalf("route has blank ID")
	}
	// delete route
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rt.Id, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rt.Id)
}

// various api tests

func TestRestGetRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	routeFileName := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/ears/v1"+tenantPath+"/routes/r100", nil)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
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
	// delete routes
	rtId := "r100"
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rtId, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rtId)
}

func TestRestGetMultipleRoutesHandler(t *testing.T) {
	Version = "v1.0.2"
	routeFileName := "testdata/simpleRoute.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/ears/v1"+tenantPath+"/routes", nil)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
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
	// delete routes
	rtId := "r100"
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rtId, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rtId)
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
	runtime := setupSimpleApi(t, "inmemory")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleFilterRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/r100", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/ears/v1"+tenantPath+"/routes", nil)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
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
	// delete remaining routes
	rtId := "f103"
	r = httptest.NewRequest(http.MethodDelete, "/ears/v1"+tenantPath+"/routes/"+rtId, nil)
	w = httptest.NewRecorder()
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	t.Logf("deleted route with id: %s", rtId)
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
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPut, "/ears/v1"+tenantPath+"/routes/badid", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
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
	runtime := setupSimpleApi(t, "inmemory")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ears/v1"+tenantPath+"/routes/fakeid", nil)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
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
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
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
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addroutebadpluginname", data)
}

func TestRestPostRouteHandlerNoSender(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	routeFileName := "testdata/simpleRouteNoSender.json"
	simpleRouteReader, err := os.Open(routeFileName)
	if err != nil {
		t.Fatalf("cannot read file: %s", err.Error())
	}
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
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
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
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
	runtime := setupSimpleApi(t, "inmemory")
	r := httptest.NewRequest(http.MethodPost, "/ears/v1"+tenantPath+"/routes", simpleRouteReader)
	runtime.apiManager.muxRouter.ServeHTTP(w, r)
	g := goldie.New(t)
	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addroutenouser", data)
}
