// Copyright 2021 Comcast Cable Communications Management, LLC
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

package db_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sort"
	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/route"
)

type RouteTestCase struct {
	tenantId    tenant.Id
	routeId     string
	routeConfig string
}

var testCases = []RouteTestCase{
	RouteTestCase{
		tenant.Id{
			OrgId: "myOrg",
			AppId: "myApp",
		},
		"test",
		`
		{
		  "userId": "mchiang",
		  "name": "myName",
		  "deliveryMode": "fire_and_forget"
		}
		`,
	},
	RouteTestCase{
		tenant.Id{
			OrgId: "myOrg",
			AppId: "myApp",
		},
		"test",
		`
		{
		  "userId": "mchiang",
		  "name": "differentName",
		  "deliveryMode": "fire_and_forget",
		  "source": {
			"someField": "blah"
		  }
		}
		`,
	},
	RouteTestCase{
		tenant.Id{
			OrgId: "myOrg2",
			AppId: "myApp2",
		},
		"test2",
		`
		{
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
		`,
	},
	RouteTestCase{
		tenantId: tenant.Id{
			OrgId: "myOrg2",
			AppId: "myApp2",
		},
		routeId: "test2",
		routeConfig: `
		{
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
		}`,
	},
	RouteTestCase{
		tenantId: tenant.Id{
			OrgId: "myOrg3",
			AppId: "myApp3",
		},
		routeId: "test3",
		routeConfig: `
		{
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
		}`,
	},
	RouteTestCase{
		tenantId: tenant.Id{
			OrgId: "myOrg3",
			AppId: "myApp3",
		},
		routeId: "test4",
		routeConfig: `
		{
		  "userId": "tgattis",
		  "name": "myName4",
		  "deliveryMode": "at_least_once",
		  "sender": {
			"plugin": "debug",
			"name": "my_debug",
			"config": {
			  "key": "value"
			}
		  }
		}`,
	},
}

func testRouteStorer(s route.RouteStorer, t *testing.T) {
	ctx := context.Background()

	//start from a clean slate
	for _, tc := range testCases {
		err := s.DeleteRoutes(ctx, tc.tenantId, []string{tc.routeId})
		if err != nil {
			t.Fatalf("DeleteAllRoutes error: %s\n", err.Error())
		}
	}
	time.Sleep(500 * time.Millisecond)

	//Test Case: tenant does not exist
	r, err := s.GetRoute(ctx, tenant.Id{OrgId: "myOrg", AppId: "myApp"}, "does_not_exist")
	if err == nil {
		t.Fatalf("Expect an error but instead get no error")
	}
	var routeNotFound *route.RouteNotFoundError
	if !errors.As(err, &routeNotFound) {
		t.Fatalf("GetRoute does_not_exist unexpected error: %s\n", err.Error())
	}

	//TestCase: set and get
	var config route.Config
	err = json.Unmarshal([]byte(testCases[0].routeConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	config.Id = testCases[0].routeId
	config.TenantId = testCases[0].tenantId

	err = s.SetRoute(ctx, config)
	if err != nil {
		t.Fatalf("SetRoute error: %s\n", err.Error())
	}

	r, err = s.GetRoute(ctx, config.TenantId, config.Id)
	if err != nil {
		t.Fatalf("GetRoute test error: %s\n", err.Error())
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

	//Test Case: route does not exist
	r, err = s.GetRoute(ctx, tenant.Id{OrgId: "myOrg", AppId: "myApp"}, "does_not_exist")
	if err == nil {
		t.Fatalf("Expect an error but instead get no error")
	}
	var routeNotFoundErr *route.RouteNotFoundError
	if !errors.As(err, &routeNotFoundErr) {
		t.Fatalf("GetRoute does_not_exist unexpected error: %s\n", err.Error())
	}

	//TestCase: update route
	err = json.Unmarshal([]byte(testCases[1].routeConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	config.Id = testCases[1].routeId
	config.TenantId = testCases[1].tenantId

	//sleep for two seconds and then update route again (to generate different create vs modified time)
	time.Sleep(2 * time.Second)

	err = s.SetRoute(ctx, config)
	if err != nil {
		t.Fatalf("SetRoute error: %s\n", err.Error())
	}

	r, err = s.GetRoute(ctx, config.TenantId, config.Id)
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

	//TestCase: set and get on a different tenant
	err = json.Unmarshal([]byte(testCases[2].routeConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	config.Id = testCases[2].routeId
	config.TenantId = testCases[2].tenantId

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

	//Test Case: try to get route on a different tenant
	r, err = s.GetRoute(ctx, tenant.Id{OrgId: "myOrg", AppId: "myApp"}, config.Id)
	if err == nil {
		t.Fatalf("Expect an error but instead get no error")
	}
	if !errors.As(err, &routeNotFoundErr) {
		t.Fatalf("GetRoute does_not_exist unexpected error: %s\n", err.Error())
	}

	//Test Case: bulk updates
	configs := make([]route.Config, 3)
	err = json.Unmarshal([]byte(testCases[3].routeConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	config.Id = testCases[3].routeId
	config.TenantId = testCases[3].tenantId
	configs[0] = config

	err = json.Unmarshal([]byte(testCases[4].routeConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	config.Id = testCases[4].routeId
	config.TenantId = testCases[4].tenantId
	configs[1] = config

	err = json.Unmarshal([]byte(testCases[5].routeConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	config.Id = testCases[5].routeId
	config.TenantId = testCases[5].tenantId
	configs[2] = config

	err = s.SetRoutes(ctx, configs)
	if err != nil {
		t.Fatalf("SetRoutes error: %s\n", err.Error())
	}

	routes, err = s.GetAllRoutes(ctx)
	if err != nil {
		t.Fatalf("GetAllRoutes error: %s\n", err.Error())
	}
	if len(routes) != 4 {
		t.Fatalf("Expect 4 routes but get %d instead\n", len(routes))
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

	//Test case: delete some routes
	err = s.DeleteRoute(ctx, testCases[0].tenantId, testCases[0].routeId)
	if err != nil {
		t.Fatalf("DeleteRoutes error: %s\n", err.Error())
	}

	err = s.DeleteRoute(ctx, testCases[2].tenantId, testCases[2].routeId)
	if err != nil {
		t.Fatalf("DeleteRoutes error: %s\n", err.Error())
	}

	r, err = s.GetRoute(ctx, testCases[2].tenantId, testCases[2].routeId)
	if err == nil {
		t.Fatalf("GetRoute Expect an error but instead get no error")
	}
	if !errors.As(err, &routeNotFoundErr) {
		t.Fatalf("GetRoute unexpected error: %s\n", err.Error())
	}

	routes, err = s.GetAllRoutes(ctx)
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
	g.AssertJson(t, "allroutes3", routes)

	routes, err = s.GetAllTenantRoutes(ctx, testCases[4].tenantId)
	if err != nil {
		t.Fatalf("GetAllTenantRoutes error: %s\n", err.Error())
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
	g.AssertJson(t, "allroutes4", routes)

	err = s.DeleteRoutes(ctx, testCases[4].tenantId, []string{testCases[4].routeId, testCases[5].routeId})
	if err != nil {
		t.Fatalf("DeleteRoute error: %s\n", err.Error())
	}

	routes, err = s.GetAllRoutes(ctx)
	if err != nil {
		t.Fatalf("GetAllRoutes error: %s\n", err.Error())
	}
	if len(routes) != 0 {
		t.Fatalf("Expect 0 routes but get %d instead\n", len(routes))
	}
}
