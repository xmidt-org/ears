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

// +build integration

package perf

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/xmidt-org/ears/pkg/tenant"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

var earsEp = flag.String("ears", "http://127.0.0.1:3000/ears", "Ears endpoint for testing")
var earthEp = flag.String("earth", "http://127.0.0.1:9090/earth", "Earth endpoint for testing")

var routeConfig = `
{
  "userId": "mchiang",
  "name": "simpleSqs",
  "receiver": {
    "plugin": "sqs",
    "name": "sqsReceiver",
    "config": {
      "queueUrl": "{{YOUR_SQS_QUEUE_1_URL_HERE}}",
      "receiverPoolSize": 10,
      "numRetries": 2
    } 
  },
  "sender": {
    "plugin": "sqs",
    "name": "sqsSender",
    "config": {
      "queueUrl": "{{YOUR_SQS_QUEUE_2_URL_HERE}}"
    } 
  },
  "deliveryMode": "whoCares"
}
`

var testConfig = `
{
  "source": {  
    "startInSec": 0,
    "num": 10,
    "ratePerSec": 0,
    "eventDistribution": { 
      "default": 50,
      "alt": 50
    },
    "events": {
      "default": {
        "mykey1": "mydata1",
        "somedata": "1234567890",
        "somemoredata": "moredata",
        "someobj": {
            "hello": "world",
            "abc": "ddd"
        }
      },
      "alt": {
        "mykey2": "mydata2",
        "somedata": "1234567890",
        "mykey2": "mydata2",
        "someobj": {
            "hello": "world",
            "abc": "ddd"
        }
      }
    },
    "type": "sqs",
    "config": {
      "queueUrl": "{{YOUR_SQS_QUEUE_1_URL_HERE}}"
    }
  },
  "destination": {
    "numEventsExpected": 10,
    "saveEvents": true,
    "type": "sqs",
    "config": {
      "queueUrl": "{{YOUR_SQS_QUEUE_2_URL_HERE}}",
      "receiverPoolSize": 40
    } 
  }
}
`

func TestPerfSQSPlugin(t *testing.T) {
	tid := tenant.Id{OrgId: "perfOrg", AppId: "perfApp"}
	routeId := "sqsPerf"
	minimumExpectedEps := 500

	setupQuota(t, *earsEp, tid, 2000)
	createRoute(t, *earsEp, tid, routeId, routeConfig)
	runPerfTest(t, *earthEp, testConfig, minimumExpectedEps)
}

func TestPerfQuotaPlugin(t *testing.T) {
	tid := tenant.Id{OrgId: "quotaOrg", AppId: "quotaApp"}
	routeId := "sqsQuota"
	minimumExpectedEps := 80

	setupQuota(t, *earsEp, tid, 100)
	createRoute(t, *earsEp, tid, routeId, routeConfig)
	runPerfTest(t, *earthEp, testConfig, minimumExpectedEps)
}

func setupQuota(t *testing.T, earsEp string, tid tenant.Id, eps int) {
	url := earsEp + "/v1/orgs/" + tid.OrgId + "/applications/" + tid.AppId + "/config"
	configTemplate := `
	{
		"quota": {
			"eventsPerSec": %d
		}
	}
	`
	config := fmt.Sprintf(configTemplate, eps)
	_, err := doHttp(http.MethodPut, url, config)
	if err != nil {
		t.Fatalf("Fail to setup tenant quota %s\n", err.Error())
	}
}

func createRoute(t *testing.T, earsEp string, tid tenant.Id, routeId string, routeConfig string) {
	url := earsEp + "/v1/orgs/" + tid.OrgId + "/applications/" + tid.AppId + "/routes/" + routeId
	_, err := doHttp(http.MethodPut, url, routeConfig)
	if err != nil {
		t.Fatalf("Fail to create route %s\n", err.Error())
	}
}

type EarthCreateJobResp struct {
	Item string `json:"item"`
}

type EarthJobReport struct {
	Status        string  `json:"status"`
	NumReceived   int     `json:"numReceived"`
	DurationInSec float64 `json:"durationInSec"`
	SrcThroughput float64 `json:"sourceThroughput"`
	DstThroughput float64 `json:"destinationThroughput"`
}

type EarthJobReportResp struct {
	Item EarthJobReport `json:"item"`
}

func runPerfTest(t *testing.T, earthEp string, testConfig string, minimumExpectedEps int) {
	url := earthEp + "/v1/jobs"

	resp, err := doHttp(http.MethodPost, url, testConfig)
	if err != nil {
		t.Fatalf("Fail to create test %s\n", err.Error())
	}

	var jobResp EarthCreateJobResp
	err = json.Unmarshal(resp, &jobResp)
	if err != nil {
		t.Fatalf("Fail to unmarshal earth job response %s\n", err.Error())
	}

	fmt.Printf("Earth job created. Id=%s\n", jobResp.Item)

	url = earthEp + "/v1/jobs/" + jobResp.Item + "/report"
	for {
		time.Sleep(time.Second)
		resp, err = doHttp(http.MethodGet, url, "")
		if err != nil {
			t.Fatalf("Error getting job report %s\n", err.Error())
		}
		var jobReport EarthJobReportResp
		err = json.Unmarshal(resp, &jobReport)
		if err != nil {
			t.Fatalf("Error unmarshal job report %s\n", err.Error())
		}
		if jobReport.Item.Status == "complete" {
			if jobReport.Item.DstThroughput < float64(minimumExpectedEps) {
				t.Fatalf("Throughput %f does not meet the expected %d\n",
					jobReport.Item.DstThroughput,
					minimumExpectedEps)
			}
			fmt.Printf("Throughput %f events/sec\n", jobReport.Item.DstThroughput)
			break
		} else if jobReport.Item.Status == "error" {
			t.Fatalf("Test failed with error status %v\n", jobReport)
			break
		} else if jobReport.Item.Status == "canceled" {
			t.Fatalf("Test failed with cancel status %v\n", jobReport)
			break
		}
	}
}

func doHttp(method string, url string, body string) ([]byte, error) {
	client := http.DefaultClient

	var req *http.Request
	var err error
	if body != "" {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("resp=%s\n", string(respData))
		return nil, fmt.Errorf("bad status code %d\n", resp.StatusCode)
	}

	return respData, nil
}
