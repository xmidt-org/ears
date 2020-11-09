package route_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/xmidt-org/ears/internal/pkg/route"

	"github.com/rs/zerolog/log"
)

//TODO: test all errors
//TODO: test hashing and validations
//TODO: test multiple routes

type (
	TestTableEntry struct {
		Name                          string
		ExpectedSourceEventCount      int
		ExpectedDestinationEventCount int
	}
)

func getTestRouteByName(t *testing.T, name string) string {
	path := filepath.Join("testdata", name+".json")
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return string(buf)
}

func TestSingleRouteTestTable(t *testing.T) {
	path := filepath.Join("testdata", "test_table.json")
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	testTable := make([]*TestTableEntry, 0)
	err = json.Unmarshal(buf, &testTable)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	for _, singleRouteTest := range testTable {
		simulateSingleRoute(t, getTestRouteByName(t, singleRouteTest.Name), singleRouteTest.ExpectedSourceEventCount, singleRouteTest.ExpectedDestinationEventCount)
	}
}

func simulateSingleRoute(t *testing.T, rstr string, expectedSourceCount, expectedDestinationCount int) {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	// init in memory routing table manager
	rtmgr := route.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	var rc route.RouteConfig
	err := json.Unmarshal([]byte(rstr), &rc)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	// add a route
	err = rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(&rc))
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
	// check if expected number of events hhave made it thhroughh input and output plugins
	if allRoutes[0].Source.GetEventCount() != expectedSourceCount {
		t.Errorf("unexpected number of produced events %d", allRoutes[0].Source.GetEventCount())
	}
	if allRoutes[0].Destination.GetEventCount() != expectedDestinationCount {
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

func TestMultiRoutes(t *testing.T) {

}

/*func TestSplitRoute(t *testing.T) {
	simulateSingleRoute(t, getTestRouteByName(t, "split_route"), 3, 6)
}*/

/*func TestDirectRoute(t *testing.T) {
	simulateSingleRoute(t, getTestRouteByName(t, "direct_route"), 1, 1)
}*/

/*func TestFilterRoute(t *testing.T) {
	simulateSingleRoute(t, getTestRouteByName(t, "filter_route"), 3, 0)
}*/

/*func TestArrayRoute(t *testing.T) {
	simulateSingleRoute(t, getTestRouteByName(t, "array_route"), 3, 0)
}*/

/*func TestWildcardRoute(t *testing.T) {
	simulateSingleRoute(t, getTestRouteByName(t, "wildcard_route"), 3, 0)
}*/

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
	routes := []string{getTestRouteByName(t, "split_route"), getTestRouteByName(t, "direct_route"), getTestRouteByName(t, "filter_route"), getTestRouteByName(t, "array_route"), getTestRouteByName(t, "wildcard_route")}
	var rc *route.RouteConfig
	for _, rstr := range routes {
		rc = new(route.RouteConfig)
		err := json.Unmarshal([]byte(rstr), rc)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		err = rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(rc))
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
	foundRoutes, err := rtmgr.GetRoutesBySourcePlugin(ctx, rc.Source)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if len(foundRoutes) != 1 {
		t.Errorf("unexpected number of routes by source %d", len(foundRoutes))
	}
	foundRoutes, err = rtmgr.GetRoutesByDestinationPlugin(ctx, rc.Destination)
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
	routes := []string{getTestRouteByName(t, "split_route"), getTestRouteByName(t, "direct_route"), getTestRouteByName(t, "filter_route"), getTestRouteByName(t, "array_route"), getTestRouteByName(t, "wildcard_route")}
	var rc *route.RouteConfig
	for _, rstr := range routes {
		rc = new(route.RouteConfig)
		err := json.Unmarshal([]byte(rstr), rc)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		err = rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(rc))
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
	routes := []string{getTestRouteByName(t, "empty_route")}
	var rc *route.RouteConfig
	for _, rstr := range routes {
		rc = new(route.RouteConfig)
		err := json.Unmarshal([]byte(rstr), rc)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		err = rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(rc))
		if err == nil {
			t.Errorf("missing error")
		}
		_, ok := err.(*route.UnknownPluginTypeError)
		if !ok {
			t.Errorf("wrong error")
		}
	}
	// validate
	err = rtmgr.Validate(ctx)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	hash := rtmgr.Hash(ctx)
	if hash != "" {
		t.Errorf("table hash not empty")
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
	routes := []string{getTestRouteByName(t, "split_route")}
	var rc *route.RouteConfig
	for _, rstr := range routes {
		rc = new(route.RouteConfig)
		err := json.Unmarshal([]byte(rstr), rc)
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		err = rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(rc))
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}
	// inject event
	allRoutes, _ := rtmgr.GetAllRoutes(ctx)
	r := allRoutes[0]
	event := route.NewEvent(ctx, r.Source, map[string]string{"foo": "bar"})
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
	/*if r.Destination.GetEventCount() != 6 {
		t.Errorf("unexpected number of events %d", r.Destination.GetEventCount())
	}*/
}
