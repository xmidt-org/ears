// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package db_test

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db/dynamo"
)

func dynamoDbConfig() config.Config {
	v := viper.New()
	v.Set("ears.storage.route.region", "us-west-2")
	v.Set("ears.storage.route.tableName", "dev.ears.routes")
	return v
}

func TestDynamoRouteStorer(t *testing.T) {
	s, err := dynamo.NewDynamoDbStorer(dynamoDbConfig())
	if err != nil {
		t.Fatalf("Error instantiate dynamodb %s\n", err.Error())
	}
	testRouteStorer(s, t)
}
