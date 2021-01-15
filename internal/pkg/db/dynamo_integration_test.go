// +build integration

package db_test

import (
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/app"
	"github.com/xmidt-org/ears/internal/pkg/db/dynamo"
	"testing"
)

func dynamoDbConfig() app.Config {
	v := viper.New()
	v.Set("ears.db.region", "us-west-2")
	v.Set("ears.db.tableName", "gears.dev.ears")
	return v
}

func TestDynamoRouteStorer(t *testing.T) {
	s, err := dynamo.NewDynamoDbStorer(dynamoDbConfig())
	if err != nil {
		t.Fatalf("Error instantiate dynamodb %s\n", err.Error())
	}
	testRouteStorer(s, t)
}
