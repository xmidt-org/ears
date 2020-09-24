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
