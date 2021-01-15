package db_test

import (
	"context"
	"encoding/json"
	"github.com/sebdah/goldie/v2"
	"sort"
	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/route"
)

var testRouteConfig1 = `
{
  "id": "test",
  "orgId": "myOrg",
  "appId": "myApp",
  "userId": "mchiang",
  "name": "myName",
  "deliveryMode": "fire_and_forget"
}
`
var testRouteConfig2 = `
{
  "id": "test",
  "orgId": "myOrg",
  "appId": "myApp",
  "userId": "mchiang",
  "name": "differentName",
  "deliveryMode": "fire_and_forget",
  "source": {
    "someField": "blah"
  }
}
`

var testRouteConfig3 = `
{
  "id": "test2",
  "orgId": "myOrg2",
  "appId": "myApp2",
  "userId": "bwolf",
  "name": "myName2",
  "deliveryMode": "at_least_once",
  "sender": {
    "plugin": "debug",
    "name": "my_debug",
    "config": {
      "key": "value"
    }
  }
}
`

var testRouteConfig4 = `
[
	{
	  "id": "test2",
	  "orgId": "myOrg2",
	  "appId": "myApp2",
	  "userId": "bwolf",
	  "name": "anotherName",
	  "deliveryMode": "at_least_once",
	  "sender": {
		"plugin": "debug",
		"name": "my_debug",
		"config": {
		  "key": "value"
		}
	  }
	},
	{
	  "id": "test3",
	  "orgId": "myOrg3",
	  "appId": "myApp3",
	  "userId": "tgattis",
	  "name": "myName3",
	  "deliveryMode": "at_least_once",
	  "sender": {
		"plugin": "debug",
		"name": "my_debug",
		"config": {
		  "key": "value"
		}
	  }
	}
]
`

func testRouteStorer(s route.RouteStorer, t *testing.T) {
	ctx := context.Background()

	//start from a clean slate
	err := s.DeleteRoutes(ctx, []string{"test", "test2", "test3"})
	if err != nil {
		t.Fatalf("DeleteAllRoutes error: %s\n", err.Error())
	}
	time.Sleep(500 * time.Millisecond)

	r, err := s.GetRoute(ctx, "does_not_exist")
	if err != nil {
		t.Fatalf("GetRoute does_not_exist error: %s\n", err.Error())
	}
	if r != nil {
		t.Fatalf("Expect empty route but instead get a route")
	}

	var config route.Config
	err = json.Unmarshal([]byte(testRouteConfig1), &config)
	if err != nil {
		t.Fatalf("Unmarshal testRouteConfig1 error: %s\n", err.Error())
	}

	err = s.SetRoute(ctx, config)
	if err != nil {
		t.Fatalf("SetRoute error: %s\n", err.Error())
	}

	r, err = s.GetRoute(ctx, "test")
	if err != nil {
		t.Fatalf("GetRoute test error: %s\n", err.Error())
	}
	if r == nil {
		t.Fatalf("GetRoute return nil")
	}

	//confirm create and modified time are set and they are equal
	if r.Created == 0 || r.Created != r.Modified {
		t.Fatalf("Unexpected create and/or modified time %d %d\n", r.Created, r.Modified)
	}
	g := goldie.New(t)

	//remove create and modified time so we can assert with goldie
	r.Created = 0
	r.Modified = 0
	g.AssertJson(t, "route", r)

	err = json.Unmarshal([]byte(testRouteConfig2), &config)
	if err != nil {
		t.Fatalf("Unmarshal testRouteConfig2 error: %s\n", err.Error())
	}

	//sleep for two seconds and then update route again (to generate different create vs modified time)
	time.Sleep(2 * time.Second)

	err = s.SetRoute(ctx, config)
	if err != nil {
		t.Fatalf("SetRoute error: %s\n", err.Error())
	}

	r, err = s.GetRoute(ctx, "test")
	if err != nil {
		t.Fatalf("GetRoute test error: %s\n", err.Error())
	}

	//confirm create and modified time are set and they are different now
	if r.Created == 0 || r.Created == r.Modified {
		t.Fatalf("Unexpected create and/or modified time %d %d\n", r.Created, r.Modified)
	}

	//remove create and modified time so we can assert with goldie
	r.Created = 0
	r.Modified = 0
	g.AssertJson(t, "route_updated", r)

	err = json.Unmarshal([]byte(testRouteConfig3), &config)
	if err != nil {
		t.Fatalf("Unmarshal testRouteConfig3 error: %s\n", err.Error())
	}

	err = s.SetRoute(ctx, config)
	if err != nil {
		t.Fatalf("SetRoute error: %s\n", err.Error())
	}

	routes, err := s.GetAllRoutes(ctx)
	if err != nil {
		t.Fatalf("GetAllRoutes error: %s\n", err.Error())
	}
	if len(routes) != 2 {
		t.Fatalf("Expect 2 routes but get %d instead\n", len(routes))
	}

	//remove create and modified time so we can assert with goldie
	for i := 0; i < len(routes); i++ {
		routes[i].Created = 0
		routes[i].Modified = 0
	}
	sort.SliceStable(routes, func(i int, j int) bool {
		return routes[i].Id < routes[j].Id
	})
	g.AssertJson(t, "allroutes", routes)

	var configs []route.Config
	err = json.Unmarshal([]byte(testRouteConfig4), &configs)
	if err != nil {
		t.Fatalf("Unmarshal testRouteConfig4 error: %s\n", err.Error())
	}

	err = s.SetRoutes(ctx, configs)
	if err != nil {
		t.Fatalf("SetRoutes error: %s\n", err.Error())
	}

	routes, err = s.GetAllRoutes(ctx)
	if err != nil {
		t.Fatalf("GetAllRoutes error: %s\n", err.Error())
	}
	if len(routes) != 3 {
		t.Fatalf("Expect 3 routes but get %d instead\n", len(routes))
	}

	//remove create and modified time so we can assert with goldie
	for i := 0; i < len(routes); i++ {
		routes[i].Created = 0
		routes[i].Modified = 0
	}
	sort.SliceStable(routes, func(i int, j int) bool {
		return routes[i].Id < routes[j].Id
	})
	g.AssertJson(t, "allroutes2", routes)

	err = s.DeleteRoutes(ctx, []string{"test", "test2"})
	if err != nil {
		t.Fatalf("DeleteRoutes error: %s\n", err.Error())
	}

	r, err = s.GetRoute(ctx, "test2")
	if err != nil {
		t.Fatalf("GetRoute test2 error: %s\n", err.Error())
	}
	if r != nil {
		t.Fatalf("Expect empty route but instead get a route")
	}

	routes, err = s.GetAllRoutes(ctx)
	if err != nil {
		t.Fatalf("GetAllRoutes error: %s\n", err.Error())
	}
	if len(routes) != 1 {
		t.Fatalf("Expect 1 routes but get %d instead\n", len(routes))
	}

	//remove create and modified time so we can assert with goldie
	for i := 0; i < len(routes); i++ {
		routes[i].Created = 0
		routes[i].Modified = 0
	}
	sort.SliceStable(routes, func(i int, j int) bool {
		return routes[i].Id < routes[j].Id
	})
	g.AssertJson(t, "allroutes3", routes)

	err = s.DeleteRoute(ctx, "test3")
	if err != nil {
		t.Fatalf("DeleteRoute test3 error: %s\n", err.Error())
	}

	routes, err = s.GetAllRoutes(ctx)
	if err != nil {
		t.Fatalf("GetAllRoutes error: %s\n", err.Error())
	}
	if len(routes) != 0 {
		t.Fatalf("Expect 0 routes but get %d instead\n", len(routes))
	}
}
