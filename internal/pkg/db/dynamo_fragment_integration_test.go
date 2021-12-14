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

//go:build integration
// +build integration

package db_test

import (
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db/dynamo"
	"testing"
)

func dynamoDbFragmentConfig() config.Config {
	v := viper.New()
	v.Set("ears.storage.fragment.region", "us-west-2")
	v.Set("ears.storage.fragment.tableName", "dev.ears.fragments")
	return v
}

func TestDynamoFragmentStorer(t *testing.T) {
	s, err := dynamo.NewDynamoDbFragmentStorer(dynamoDbFragmentConfig())
	if err != nil {
		t.Fatalf("Error instantiate dynamodb %s\n", err.Error())
	}
	testFragmentStorer(s, t)
}
