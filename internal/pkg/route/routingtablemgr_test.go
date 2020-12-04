package route_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/xmidt-org/ears/internal/pkg/route"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	TestTableSingleEntry struct {
		Name                          string
		ExpectedSourceEventCount      int
		ExpectedDestinationEventCount int
	}
	TestTableMultiEntry struct {
		Name                    string
		Routes                  []string
		InputPluginEventCounts  map[string]int
		OutputPluginEventCounts map[string]int
		OutputPluginEvents      map[string]string
	}
)

func getTestRouteByName(t *testing.T, name string) *route.RouteConfig {
	path := filepath.Join("testdata/routes", name+".json")
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("unknown test route file: " + err.Error())
	}
	var rc route.RouteConfig
	err = json.Unmarshal(buf, &rc)
	if err != nil {
		t.Fatalf("unknown test route " + name + ": " + err.Error())
	}
	return &rc
}

func getTestEventByName(t *testing.T, name string) interface{} {
	path := filepath.Join("testdata/events", name+".json")
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("unknown test event file: " + err.Error())
	}
	var evt interface{}
	err = json.Unmarshal(buf, &evt)
	if err != nil {
		t.Fatalf("unknown test event " + name + ": " + err.Error())
	}
	return evt
}

func TestSingleRouteTestTable(t *testing.T) {
	path := filepath.Join("testdata", "test_table_single_route.json")
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	testTable := make([]*TestTableSingleEntry, 0)
	err = json.Unmarshal(buf, &testTable)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	for _, singleRouteTest := range testTable {
		simulateSingleRoute(t, singleRouteTest)
	}
}

func TestMultiRouteTestTable(t *testing.T) {
	path := filepath.Join("testdata", "test_table_multi_route.json")
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	testTable := make([]*TestTableMultiEntry, 0)
	err = json.Unmarshal(buf, &testTable)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	for _, multiRouteTest := range testTable {
		simulateMultiRoutes(t, multiRouteTest)
	}
}

func simulateSingleRoute(t *testing.T, data *TestTableSingleEntry) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	// init in memory routing table manager
	rtmgr := route.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	// add a route
	rc := getTestRouteByName(t, data.Name)
	err := rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(rc))
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	// check route
	if rtmgr.GetRouteCount(ctx) != 1 {
		t.Errorf("routing table doesn't have expected entry")
	}
	allRoutes, err := rtmgr.GetAllRoutes(ctx)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(allRoutes) != 1 {
		t.Errorf("routing table doesn't have expected entry")
	}
	// give plugins some time to do their thing
	time.Sleep(time.Duration(1000) * time.Millisecond)
	// check if expected number of events have made it through input and output plugins
	if allRoutes[0].Source.GetEventCount() != data.ExpectedSourceEventCount {
		t.Errorf("unexpected number of produced events %d", allRoutes[0].Source.GetEventCount())
	}
	if allRoutes[0].Destination.GetEventCount() != data.ExpectedDestinationEventCount {
		t.Errorf("unexpected number of consumed events %d", allRoutes[0].Destination.GetEventCount())
	}
	// cleanup routes
	err = rtmgr.ReplaceAllRoutes(ctx, nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
	}
}

func simulateMultiRoutes(t *testing.T, data *TestTableMultiEntry) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	// init in memory routing table manager
	rtmgr := route.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	// add a routes
	for _, name := range data.Routes {
		/*rstr := getTestRouteByName(t, name)
		var rc *route.RouteConfig
		rc = new(route.RouteConfig)
		rc.Source = new(route.Plugin)
		rc.Destination = new(route.Plugin)
		err := json.Unmarshal([]byte(rstr), rc)
		if err != nil {
			t.Errorf(err.Error())
			return
		}*/
		route := route.NewRouteFromRouteConfig(getTestRouteByName(t, name))
		err := rtmgr.AddRoute(ctx, route)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}
	// validate
	err := rtmgr.Validate(ctx)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	hash := rtmgr.Hash(ctx)
	if hash == "" {
		t.Errorf("empty table hash")
	}
	// check routes
	if rtmgr.GetRouteCount(ctx) != len(data.Routes) {
		t.Errorf("routing table doesn't have expected number of entries %d", rtmgr.GetRouteCount(ctx))
	}
	allRoutes, err := rtmgr.GetAllRoutes(ctx)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(allRoutes) != len(data.Routes) {
		t.Errorf("unexpected number of routes %d", len(allRoutes))
		return
	}
	// give plugins some time to do their thing
	time.Sleep(time.Duration(1000) * time.Millisecond)
	// check if expected number of events have made it through input and output plugins
	for k, v := range data.InputPluginEventCounts {
		for _, r := range allRoutes {
			if r.Source.GetConfig().Name == k {
				if r.Source.GetEventCount() != v {
					t.Errorf("unexpected number of produced events %d for %s (expected %d)", r.Source.GetEventCount(), k, v)
				}
			}
		}
	}
	for k, v := range data.OutputPluginEventCounts {
		for _, r := range allRoutes {
			if r.Destination.GetConfig().Name == k {
				if r.Destination.GetEventCount() != v {
					t.Errorf("unexpected number of consumed events %d for %s (expected %d)", r.Destination.GetEventCount(), k, v)
				}
				if data.OutputPluginEvents != nil {
					expectedEvent := getTestEventByName(t, data.OutputPluginEvents[k])
					if expectedEvent != nil {
						lastEvent := r.Destination.GetLastEvent()
						if lastEvent == nil {
							t.Errorf("no event received for %s", k)
						} else if !reflect.DeepEqual(expectedEvent, lastEvent.Payload) {
							t.Errorf("unexpected event payload for %s", k)
						}
					}
				}
			}
		}
	}
	// cleanup routes
	err = rtmgr.ReplaceAllRoutes(ctx, nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
	}
}

func TestGetRoutesBy(t *testing.T) {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	// init in memory routing table manager
	rtmgr := route.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	// add a routes
	routes := []*route.RouteConfig{getTestRouteByName(t, "split_route"), getTestRouteByName(t, "direct_route"), getTestRouteByName(t, "filter_route"), getTestRouteByName(t, "array_route"), getTestRouteByName(t, "wildcard_route")}
	for _, rc := range routes {
		err := rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(rc))
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}
	// validate
	err := rtmgr.Validate(ctx)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	hash := rtmgr.Hash(ctx)
	if hash == "" {
		t.Errorf("empty table hash")
	}
	// check routes
	if rtmgr.GetRouteCount(ctx) != 5 {
		t.Errorf("routing table doesn't have expected number of entries %d", rtmgr.GetRouteCount(ctx))
	}
	foundRoutes, err := rtmgr.GetRoutesBySourcePlugin(ctx, routes[4].Source)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(foundRoutes) != 1 {
		t.Errorf("unexpected number of routes by source %d", len(foundRoutes))
	}
	foundRoutes, err = rtmgr.GetRoutesByDestinationPlugin(ctx, routes[4].Destination)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(foundRoutes) != 5 {
		t.Errorf("unexpected number of routes by destination %d", len(foundRoutes))
	}
	err = rtmgr.RemoveRoute(ctx, foundRoutes[0])
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if rtmgr.GetRouteCount(ctx) != 4 {
		t.Errorf("unexpected number of routes %d", len(foundRoutes))
	}
}

func TestGetPluginsBy(t *testing.T) {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	// init in memory routing table manager
	rtmgr := route.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	// add a routes
	routes := []*route.RouteConfig{getTestRouteByName(t, "split_route"), getTestRouteByName(t, "direct_route"), getTestRouteByName(t, "filter_route"), getTestRouteByName(t, "array_route"), getTestRouteByName(t, "wildcard_route")}
	for _, rc := range routes {
		err := rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(rc))
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}
	// validate
	err := rtmgr.Validate(ctx)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	hash := rtmgr.Hash(ctx)
	if hash == "" {
		t.Errorf("empty table hash")
	}
	// check plugins
	pmgr := route.GetIOPluginManager(ctx)
	plugins, err := pmgr.GetAllPlugins(ctx)
	if err != nil {
		t.Errorf("unexpcted error " + err.Error())
	}
	if len(plugins) != 5 {
		t.Errorf("unexpcted number of plugins %d", len(plugins))
	}
	// check plugins by mode
	plugins, err = pmgr.GetPlugins(ctx, "input", "")
	if err != nil {
		t.Errorf("unexpcted error " + err.Error())
	}
	if len(plugins) != 4 {
		t.Errorf("unexpcted number of input plugins %d", len(plugins))
	}
	plugins, err = pmgr.GetPlugins(ctx, "output", "")
	if err != nil {
		t.Errorf("unexpcted error " + err.Error())
	}
	if len(plugins) != 1 {
		t.Errorf("unexpcted number of output plugins %d", len(plugins))
	}
	// check plugins by type
	plugins, err = pmgr.GetPlugins(ctx, "", "debug")
	if err != nil {
		t.Errorf("unexpcted error " + err.Error())
	}
	if len(plugins) != 5 {
		t.Errorf("unexpcted number of debug plugins %d", len(plugins))
	}
	// remove all routes
	err = rtmgr.ReplaceAllRoutes(ctx, nil)
	if err != nil {
		t.Errorf("unexpcted error " + err.Error())
	}
	// and ensure all plugins have been withdrawn
	plugins, err = pmgr.GetAllPlugins(ctx)
	if err != nil {
		t.Errorf("unexpcted error " + err.Error())
	}
	if len(plugins) > 0 {
		t.Errorf("found residual plugins %d", len(plugins))
	}
}

func TestErrors(t *testing.T) {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	// init in memory routing table manager
	rtmgr := route.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	// add a routes
	err := rtmgr.AddRoute(ctx, nil)
	if err == nil {
		t.Errorf("no error")
	} else {
		fmt.Printf("error %s\n", err.Error())
	}
	// validate
	err = rtmgr.Validate(ctx)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	hash := rtmgr.Hash(ctx)
	if hash == "" {
		t.Errorf("table hash empty")
	}
}

func TestDoSync(t *testing.T) {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	// init in memory routing table manager
	rtmgr := route.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	// add a routes
	routes := []*route.RouteConfig{getTestRouteByName(t, "split_route_zero_rounds")}
	for _, rc := range routes {
		err := rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(rc))
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}
	allRoutes, _ := rtmgr.GetAllRoutes(ctx)
	r := allRoutes[0]
	event := route.NewEvent(ctx, r.Source, map[string]interface{}{"foo": "bar"})
	// inject events
	err := r.Source.DoSync(ctx, event)
	if err != nil {
		t.Errorf(err.Error())
	}
	err = r.Destination.DoSync(ctx, event)
	if err != nil {
		t.Errorf(err.Error())
	}
	err = r.FilterChain.Filters[0].DoSync(ctx, event)
	if err != nil {
		t.Errorf(err.Error())
	}
	// give plugins some time to do their thing
	time.Sleep(time.Duration(500) * time.Millisecond)
	// check event count (two of thhe events are doubled by the event splitter)
	if r.Destination.GetEventCount() != 5 {
		t.Errorf("unexpected number of events %d", r.Destination.GetEventCount())
	}
}
