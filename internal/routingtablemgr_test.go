package internal_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/xmidt-org/ears/internal"
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
)

var (
	ctx = context.Background()
)

func simulateSingleRoute(t *testing.T, route string, expectedSourceCount, expectedDestinationCount int) {
	ctx := context.Background()
	// init in memory routing table manager
	rtmgr := internal.NewInMemoryRoutingTableManager()
	if rtmgr.GetRouteCount(ctx) != 0 {
		t.Errorf("routing table not empty")
		return
	}
	var rc internal.RouteConfig
	err := json.Unmarshal([]byte(route), &rc)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	// add a route
	err = rtmgr.AddRoute(ctx, internal.NewRoutingTableEntryFromRouteConfig(&rc))
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
	fmt.Printf("%s\n", internal.GetIOPluginManager(ctx).String())*/
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
