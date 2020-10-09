package internal_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/xmidt-org/ears/internal"
)

var (
	ROUTE_1 = `
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
	ROUTE_2 = `
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
		"deliveryMode" : "at_least_once"
	}
	`
)

var (
	ctx = context.Background()
)

//TODO: merge test code

func TestSplitRoute(t *testing.T) {
	ctx := context.Background()
	// init in memory routing table manager
	rtmgr := internal.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	var rte internal.RoutingTableEntry
	err := json.Unmarshal([]byte(ROUTE_1), &rte)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	// add a route
	err = rtmgr.AddRoute(ctx, &rte)
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
	fmt.Printf("ROUTING TABLE:\n")
	fmt.Printf("%s\n", rtmgr.String())
	fmt.Printf("INPUT PLUGINS:\n")
	fmt.Printf("%s\n", internal.GetInputPluginManager(ctx).String())
	fmt.Printf("OUTPUT PLUGINS:\n")
	fmt.Printf("%s\n", internal.GetOutputPluginManager(ctx).String())
	time.Sleep(time.Duration(2000) * time.Millisecond)
	if allRoutes[0].Source.EventCount != 3 {
		t.Errorf("unexpected number of produced events %d", allRoutes[0].Source.EventCount)
	}
	if allRoutes[0].Destination.EventCount != 6 {
		t.Errorf("unexpected number of consumed events %d", allRoutes[0].Destination.EventCount)
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

func TestDirectRoute(t *testing.T) {
	ctx := context.Background()
	// init in memory routing table manager
	rtmgr := internal.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	var rte internal.RoutingTableEntry
	err := json.Unmarshal([]byte(ROUTE_2), &rte)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	// add a route
	err = rtmgr.AddRoute(ctx, &rte)
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
	fmt.Printf("ROUTING TABLE:\n")
	fmt.Printf("%s\n", rtmgr.String())
	fmt.Printf("INPUT PLUGINS:\n")
	fmt.Printf("%s\n", internal.GetInputPluginManager(ctx).String())
	fmt.Printf("OUTPUT PLUGINS:\n")
	fmt.Printf("%s\n", internal.GetOutputPluginManager(ctx).String())
	time.Sleep(time.Duration(2000) * time.Millisecond)
	if allRoutes[0].Source.EventCount != 3 {
		t.Errorf("unexpected number of produced events %d", allRoutes[0].Source.EventCount)
	}
	if allRoutes[0].Destination.EventCount != 3 {
		t.Errorf("unexpected number of consumed events %d", allRoutes[0].Destination.EventCount)
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
