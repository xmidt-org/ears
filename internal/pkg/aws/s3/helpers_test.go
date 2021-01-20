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
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestParseUrl(t *testing.T) {
	s := client{}
	var testCases = []struct {
		name        string
		url         string
		bucket      string
		path        string
		expectedErr error
	}{
		{
			name:        "expected",
			url:         "s3://bucket/path/that/is/long",
			bucket:      "bucket",
			path:        "path/that/is/long",
			expectedErr: nil,
		},
		{
			name:        "http test",
			url:         "http://serverinformation/bucket/path/that/is/long",
			bucket:      "",
			path:        "",
			expectedErr: errors.New("unsupported protocol"),
		},
		{
			name:        "format test",
			url:         "serverinformation/bucket/path/that/is/long",
			bucket:      "",
			path:        "",
			expectedErr: errors.New("poorly formatted url"),
		},
	}

	for i, test := range testCases {
		bucket, path, err := s.parseUrl(test.url, 2)
		if bucket != test.bucket {
			t.Errorf("#%d - expected bucket '%s' != '%s'\n", i, test.bucket, bucket)
		}

		if path != test.path {
			t.Errorf("#%d - expected path '%s' != '%s'\n", i, test.path, path)
		}

		switch {
		case err == test.expectedErr:
			// all good
		case err == nil && test.expectedErr != nil,
			err != nil && test.expectedErr == nil,
			err.Error() != test.expectedErr.Error():
			t.Errorf("%d - expected error '%s' != '%s'\n", i, test.expectedErr, err)
		}

		// fmt.Println(i, test, bucket, path, err)
	}
}

func TestParseUrlErrors(t *testing.T) {
	s := client{}
	var testCases = []struct {
		name        string
		url         string
		isTemporary bool
	}{
		{
			name:        "permanent error",
			url:         "http://serverinformation/bucket/path/that/is/long",
			isTemporary: false,
		},
	}

	for i, test := range testCases {
		_, _, err := s.parseUrl(test.url, 2)

		if test.isTemporary != IsTemporary(err) {
			t.Errorf("%d - expected isTemporary '%t' != '%t'\n", i, test.isTemporary, IsTemporary(err))
		}
		// fmt.Println(i, test, bucket, path, err)
	}
}

func TestGetSvc(t *testing.T) {
	c1 := client{}
	svc, err := c1.getSvc()

	if err != nil {
		t.Errorf("c1.getSvc error: " + err.Error())
	}

	if svc == nil {
		t.Errorf("c1.getSvc returned nil service")
	}

	var fakeSvc = s3.S3{}
	c2, err := New(WithService(&fakeSvc))
	if err != nil {
		t.Errorf("new service returned error: " + err.Error())
	}

	svc, err = c2.getSvc()
	if err != nil {
		t.Errorf("c2.getSvc error: " + err.Error())
	}

	if svc == nil {
		t.Errorf("c2.getSvc returned nil service")
	}

	if *svc != fakeSvc {
		t.Errorf("c2.getSvc did not return original fakeSvc")
	}

}

func TestGetConfig(t *testing.T) {
	c1 := client{}
	cfg1, err := c1.getConfig()

	if err != nil {
		t.Errorf("c1.getConfig error: " + err.Error())
	}

	if cfg1 == nil {
		t.Errorf("c1.getConfig returned nil config")
	}

	fakeCfg := aws.Config{}
	c2, err := New(WithConfig(&fakeCfg))
	if err != nil {
		t.Errorf("new service returned error: " + err.Error())
	}

	cfg2, err := c2.getConfig()

	if err != nil {
		t.Errorf("c2.getConfig error: " + err.Error())
	}

	if cfg2 == nil {
		t.Errorf("c2.getConfig returned nil config")
	}

	if cfg2 != &fakeCfg {
		t.Errorf("c2.getConfig did not return original fakeCfg")
	}

}
