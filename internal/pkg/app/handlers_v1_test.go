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
	"encoding/json"
	goldie "github.com/sebdah/goldie/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/version", nil)
	api, err := NewAPIManager(&DefaultRoutingTableManager{})
	if err != nil {
		t.Fatalf("Fail to instaniate api manager: %s\n", err.Error())
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

func TestAddRouteHandler(t *testing.T) {
	Version = "v1.0.2"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/v1/routes/testroute", nil)
	api, err := NewAPIManager(&DefaultRoutingTableManager{})
	if err != nil {
		t.Fatalf("Fail to instaniate api manager: %s\n", err.Error())
	}
	api.addRouteHandler(w, r)
	g := goldie.New(t)

	var data interface{}
	err = json.Unmarshal(w.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("cannot unmarshal response %s into json %s", string(w.Body.Bytes()), err.Error())
	}
	g.AssertJson(t, "addroute", data)
}
