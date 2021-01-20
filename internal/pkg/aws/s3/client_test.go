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

package s3

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func TestGetObjectErrors(t *testing.T) {
	t.SkipNow()
	c, err := New()
	if err != nil {
		t.Error(err)
	}

	data, err := c.GetObject("s3://badbucket/gears/config/loggerapi/gears-dev-service-loggerapi-v0-2.yaml")
	fmt.Println(data)
	fmt.Println("--------------------------------------------------------------")

	if IsTemporary(err) {
		t.Errorf("Missing file should result in IsTemporary(err) == false")
	} else {
		fmt.Println("Request marked as not temporary")
	}

	re, ok := err.(requestError)

	fmt.Println("OK", ok)
	fmt.Printf("Error string: %s\n", re)
	fmt.Printf("Error obj: %+v\n", re)
	fmt.Println("--------------------------------------------------------------")
	fmt.Printf("My Cause: %+v\n", re.Cause())
	fmt.Println("--------------------------------------------------------------")
	fmt.Printf("Errors Cause: %+v\n", errors.Cause(err))

}

func TestPutObject(t *testing.T) {
	// We don't actually want to push data to s3 each time.  Skip until
	// a better "tags" testing system is implemented
	t.Skip()
	c, err := New()
	if err != nil {
		t.Error(err)
	}

	path := "s3://xas-tests/github.comcast.com/VariousArtists/common/aws/s3/"

	testCases := []struct {
		fileName string
		data     string
	}{
		{
			fileName: "f1.json",
			data:     `{"one":1}`,
		},
	}

	for i, tc := range testCases {
		url := path + tc.fileName
		err := c.PutObject(url, tc.data)
		if err != nil {
			t.Error(err)
		}

		data, err := c.GetObject(url)

		if !cmp.Equal(tc.data, data) {
			t.Errorf("#%d Data put and data received are not the same: %v != %v", i, tc.data, data)
		}
	}
}
