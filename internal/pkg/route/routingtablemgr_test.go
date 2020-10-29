package route_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/xmidt-org/ears/internal/pkg/route"

	"github.com/rs/zerolog/log"
)

var (
	SPLIT_ROUTE = `
	{
		"orgId" : "comcast",
		"appId" : "xfi",
		"userId" : "boris",
		"source" : {
			"type" : "debug",
			"params" :
			{
				"rounds" : 3,
				"intervalMS" : 250,
				"payload" : {
					"foo" : "bar"
				}
			}
		},
		"destination" : {
			"type" : "debug",
			"params" : {}
		},
		"filterChain" : {
			"filters": 
			[
				{
					"type" : "match",
					"params" : {
						"pattern" : {
							"foo" : "bar"
						}
					}
				},
				{
					"type" : "filter",
					"params" : {
						"pattern" : {
							"hello" : "world"
						}
					}
				},
				{
					"type" : "pass",
					"params" : {}
				},
				{
					"type" : "split",
					"params" : {}
				},
				{
					"type" : "transform",
					"params" : {}
				}
			]
		},
		"deliveryMode" : "at_least_once"
	}
	`
	DIRECT_ROUTE = `
	{
		"orgId" : "comcast",
		"appId" : "xfi",
		"userId" : "boris",
		"source" : {
			"type" : "debug",
			"params" :
			{
				"rounds" : 1,
				"intervalMS" : 250,
				"payload" : {
					"foo" : "bar"
				}
			}
		},
		"destination" : {
			"type" : "debug",
			"params" : {}
		},
		"deliveryMode" : "at_least_once"
	}
	`

	FILTER_ROUTE = `
	{
		"orgId" : "comcast",
		"appId" : "xfi",
		"userId" : "boris",
		"source" : {
			"type" : "debug",
			"params" :
			{
				"rounds" : 3,
				"intervalMS" : 250,
				"payload" : {
					"foo" : "bar"
				}
			}
		},
		"destination" : {
			"type" : "debug",
			"params" : {}
		},
		"filterChain" : {
			"filters": 
			[
				{
					"type" : "match",
					"params" : {
						"pattern" : {
							"foo" : "bar"
						}
					}
				},
				{
					"type" : "filter",
					"params" : {
						"pattern" : {
							"foo" : "bar"
						}
					}
				}
			]
		},
		"deliveryMode" : "at_least_once"
	}
	`

	ARRAY_ROUTE = `
	{
		"orgId" : "comcast",
		"appId" : "xfi",
		"userId" : "boris",
		"source" : {
			"type" : "debug",
			"params" :
			{
				"rounds" : 3,
				"intervalMS" : 250,
				"payload" : [{"foo" : "bar"}]
			}
		},
		"destination" : {
			"type" : "debug",
			"params" : {}
		},
		"filterChain" : {
			"filters": 
			[
				{
					"type" : "match",
					"params" : {
						"pattern" : [{"foo" : "bar"}]
					}
				},
				{
					"type" : "filter",
					"params" : {
						"pattern" : [{"foo" : "bar"}]
					}
				}
			]
		},
		"deliveryMode" : "at_least_once"
	}
	`

	WILDCARD_ROUTE = `
	{
		"orgId" : "comcast",
		"appId" : "xfi",
		"userId" : "boris",
		"source" : {
			"type" : "debug",
			"params" :
			{
				"rounds" : 3,
				"intervalMS" : 333,
				"payload" : {"foo" : "bar"}
			}
		},
		"destination" : {
			"type" : "debug",
			"params" : {}
		},
		"filterChain" : {
			"filters": 
			[
				{
					"type" : "match",
					"params" : {
						"pattern" : {"foo" : "b*"}
					}
				},
				{
					"type" : "match",
					"params" : {
						"pattern" : {"foo" : "bar||baz"}
					}
				},
				{
					"type" : "filter",
					"params" : {
						"pattern" : {"foo" : "bar"}
					}
				}
			]
		},
		"deliveryMode" : "at_least_once"
	}
	`

	EMPTY_ROUTE = `
	{
		"orgId" : "comcast",
		"appId" : "xfi",
		"userId" : "boris",
		"deliveryMode" : "at_least_once",
		"source" : {},
		"destination" : {},
		"filterChain" : {}
	}
	`
)

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
	/*fmt.Printf("ROUTING TABLE:\n")
	fmt.Printf("%s\n", rtmgr.String())
	fmt.Printf("PLUGINS:\n")
	fmt.Printf("%s\n", route.GetIOPluginManager(ctx).String())*/
	time.Sleep(time.Duration(2000) * time.Millisecond)
	if allRoutes[0].Source.GetEventCount() != expectedSourceCount {
		t.Errorf("unexpected number of produced events %d", allRoutes[0].Source.GetEventCount())
	}
	if allRoutes[0].Destination.GetEventCount() != expectedDestinationCount {
		t.Errorf("unexpected number of consumed events %d", allRoutes[0].Destination.GetEventCount())
	}
	err = rtmgr.ReplaceAllRoutes(ctx, nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
	}
}

func TestSplitRoute(t *testing.T) {
	simulateSingleRoute(t, SPLIT_ROUTE, 3, 6)
}

func TestDirectRoute(t *testing.T) {
	simulateSingleRoute(t, DIRECT_ROUTE, 1, 1)
}

func TestFilterRoute(t *testing.T) {
	simulateSingleRoute(t, FILTER_ROUTE, 3, 0)
}

func TestArrayRoute(t *testing.T) {
	simulateSingleRoute(t, ARRAY_ROUTE, 3, 0)
}

func TestWildcardRoute(t *testing.T) {
	simulateSingleRoute(t, WILDCARD_ROUTE, 3, 0)
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
	routes := []string{SPLIT_ROUTE, DIRECT_ROUTE, FILTER_ROUTE, ARRAY_ROUTE, WILDCARD_ROUTE}
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
	routes := []string{SPLIT_ROUTE, DIRECT_ROUTE, FILTER_ROUTE, ARRAY_ROUTE, WILDCARD_ROUTE}
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
	plugins, err = pmgr.GetPlugins(ctx, "", "debug")
	if err != nil {
		t.Errorf("unexpcted error " + err.Error())
	}
	if len(plugins) != 5 {
		t.Errorf("unexpcted number of debug plugins %d", len(plugins))
	}
	//TODO: check for plugin withdrawal
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
	routes := []string{EMPTY_ROUTE}
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
	routes := []string{SPLIT_ROUTE}
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
