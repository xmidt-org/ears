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
