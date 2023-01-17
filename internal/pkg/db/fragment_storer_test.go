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
	"github.com/xmidt-org/ears/pkg/fragments"
	"github.com/xmidt-org/ears/pkg/tenant"
	"testing"
	"time"

	"github.com/xmidt-org/ears/pkg/route"
)

type FragmentTestCase struct {
	tenantId       tenant.Id
	fragmentId     string
	fragmentConfig string
}

var fragmentTestCases = []FragmentTestCase{
	FragmentTestCase{
		tenant.Id{
			OrgId: "myOrg",
			AppId: "myApp",
		},
		"mySqsSender",
		`
		{
		  "plugin": "sqs",
		  "fragmentName": "mySqsSender",
		  "config": {}
		}
		`,
	},
	FragmentTestCase{
		tenant.Id{
			OrgId: "myOrg",
			AppId: "myApp",
		},
		"mySqsSender",
		`
		{
		  "plugin": "sqs",
		  "fragmentName": "mySqsSender",
		  "config": { "foo" : "bar" }
		}
		`,
	},
	FragmentTestCase{
		tenant.Id{
			OrgId: "yourOrg",
			AppId: "yourApp",
		},
		"mySqsSender",
		`
		{
		  "plugin": "sqs",
		  "fragmentName": "mySqsSender",
		  "config": { }
		}
		`,
	},
	FragmentTestCase{
		tenant.Id{
			OrgId: "yourOrg",
			AppId: "yourApp",
		},
		"yourSqsSender",
		`
		{
		  "plugin": "sqs",
		  "fragmentName": "yourSqsSender",
		  "config": { }
		}
		`,
	},
	FragmentTestCase{
		tenant.Id{
			OrgId: "yourOrg",
			AppId: "yourApp",
		},
		"herSqsSender",
		`
		{
		  "plugin": "sqs",
		  "fragmentName": "herSqsSender",
		  "config": { }
		}
		`,
	},
}

func testFragmentStorer(s fragments.FragmentStorer, t *testing.T) {
	ctx := context.Background()

	//start from a clean slate
	for _, tc := range fragmentTestCases {
		err := s.DeleteFragments(ctx, tc.tenantId, []string{tc.fragmentId})
		if err != nil {
			t.Fatalf("DeleteAllFragments error: %s\n", err.Error())
		}
	}
	time.Sleep(500 * time.Millisecond)

	//Test Case: tenant does not exist
	_, err := s.GetFragment(ctx, tenant.Id{OrgId: "myOrg", AppId: "myApp"}, "does_not_exist")
	if err == nil {
		t.Fatalf("Expect an error but instead get no error")
	}
	var fragmentNotFound *fragments.FragmentNotFoundError
	if !errors.As(err, &fragmentNotFound) {
		t.Fatalf("GetFragment does_not_exist unexpected error: %s\n", err.Error())
	}

	//TestCase: set and get
	var config route.PluginConfig
	err = json.Unmarshal([]byte(fragmentTestCases[0].fragmentConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}

	err = s.SetFragment(ctx, fragmentTestCases[0].tenantId, config)
	if err != nil {
		t.Fatalf("SetFragment error: %s\n", err.Error())
	}

	f, err := s.GetFragment(ctx, fragmentTestCases[0].tenantId, config.FragmentName)
	if err != nil {
		t.Fatalf("GetFragment test error: %s\n", err.Error())
	}
	g := goldie.New(t)

	g.AssertJson(t, "fragment", f)

	//Test Case: route does not exist
	_, err = s.GetFragment(ctx, tenant.Id{OrgId: "myOrg", AppId: "myApp"}, "does_not_exist")
	if err == nil {
		t.Fatalf("Expect an error but instead get no error")
	}
	var fragmentNotFoundErr *fragments.FragmentNotFoundError
	if !errors.As(err, &fragmentNotFoundErr) {
		t.Fatalf("GetFragment does_not_exist unexpected error: %s\n", err.Error())
	}

	//TestCase: update route
	err = json.Unmarshal([]byte(fragmentTestCases[1].fragmentConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}

	//sleep for two seconds and then update route again (to generate different create vs modified time)
	time.Sleep(2 * time.Second)

	err = s.SetFragment(ctx, fragmentTestCases[1].tenantId, config)
	if err != nil {
		t.Fatalf("SetFragment error: %s\n", err.Error())
	}

	f, err = s.GetFragment(ctx, fragmentTestCases[1].tenantId, config.FragmentName)
	if err != nil {
		t.Fatalf("GetFragment test error: %s\n", err.Error())
	}

	g.AssertJson(t, "fragment_updated", f)

	//TestCase: set and get on a different tenant
	err = json.Unmarshal([]byte(fragmentTestCases[2].fragmentConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	err = s.SetFragment(ctx, fragmentTestCases[2].tenantId, config)
	if err != nil {
		t.Fatalf("SetFragment error: %s\n", err.Error())
	}

	fragments, err := s.GetAllFragments(ctx)
	if err != nil {
		t.Fatalf("GetAllFragments error: %s\n", err.Error())
	}
	if len(fragments) != 2 {
		t.Fatalf("Expect 2 fragments but got %d instead\n", len(fragments))
	}

	//Test Case: bulk updates
	configs := make([]route.PluginConfig, 2)
	err = json.Unmarshal([]byte(fragmentTestCases[3].fragmentConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	configs[0] = config

	err = json.Unmarshal([]byte(fragmentTestCases[4].fragmentConfig), &config)
	if err != nil {
		t.Fatalf("Unmarshal error: %s\n", err.Error())
	}
	configs[1] = config

	err = s.SetFragments(ctx, fragmentTestCases[3].tenantId, configs)
	if err != nil {
		t.Fatalf("SetFragments error: %s\n", err.Error())
	}

	fragments, err = s.GetAllFragments(ctx)
	if err != nil {
		t.Fatalf("GetAllFragments error: %s\n", err.Error())
	}
	if len(fragments) != 4 {
		t.Fatalf("Expect 4 routes but get %d instead\n", len(fragments))
	}

	//Test case: delete some routes
	err = s.DeleteFragment(ctx, fragmentTestCases[0].tenantId, fragmentTestCases[0].fragmentId)
	if err != nil {
		t.Fatalf("DeleteFragment error: %s\n", err.Error())
	}

	err = s.DeleteFragment(ctx, fragmentTestCases[2].tenantId, fragmentTestCases[2].fragmentId)
	if err != nil {
		t.Fatalf("DeleteFRagment error: %s\n", err.Error())
	}

	_, err = s.GetFragment(ctx, fragmentTestCases[2].tenantId, fragmentTestCases[2].fragmentId)
	if err == nil {
		t.Fatalf("GetFragment expect an error but instead get no error")
	}
	if !errors.As(err, &fragmentNotFoundErr) {
		t.Fatalf("GetFragment unexpected error: %s\n", err.Error())
	}

	fragments, err = s.GetAllFragments(ctx)
	if err != nil {
		t.Fatalf("GetAllFragments error: %s\n", err.Error())
	}
	if len(fragments) != 2 {
		t.Fatalf("Expect 2 fragments but get %d instead\n", len(fragments))
	}
}
