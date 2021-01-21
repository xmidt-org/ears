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

// Package s3 provides simplified aws s3 functionality in order to reduce the
// amount of duplicated s3 code we have across all of our Go projects
package s3

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

// getSvc will return the service stored in s3, or it will create
// a new one and return it (but won't store it in s3)
func (s *client) getSvc() (*s3.S3, error) {
	if s.svc != nil {
		return s.svc, nil
	}

	cfg, err := s.getConfig()

	if err != nil {
		return nil, err
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	return s3.New(sess), nil
}

func DefaultConfig() (*aws.Config, error) {
	return aws.NewConfig(), nil
}

func (s *client) getConfig() (*aws.Config, error) {
	if s.cfg != nil {
		return s.cfg, nil
	}

	// cfg, err := external.LoadDefaultAWSConfig()
	// if err != nil {
	// 	return nil, baseError{cause: errors.Wrap(err, "load_default_aws_config")}
	// }

	return DefaultConfig()
}

func (s *client) parseUrl(url string, minNumParts int) (string, string, error) {

	parts := strings.SplitN(url, "://", 2)

	if 2 != len(parts) {
		return "", "", parameterError{
			baseError{cause: errors.New("poorly formatted url")},
			"url",
			url,
		}

	}

	if parts[0] != "s3" {
		return "", "", parameterError{
			baseError{cause: errors.New("unsupported protocol")},
			"protocol",
			parts[0],
		}

	}
	parts = strings.SplitN(parts[1], "/", 2)
	if minNumParts > len(parts) {
		return "", "", parameterError{
			baseError{cause: errors.New("poorly formatted url")},
			"url",
			url,
		}
	}

	prefix := ""
	if 2 == len(parts) {
		prefix = parts[1]
	}

	return parts[0], prefix, nil
}
