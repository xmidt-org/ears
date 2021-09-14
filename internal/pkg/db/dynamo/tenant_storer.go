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

package dynamo

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/semconv/v1.4.0"
	"time"
)

type TenantStorer struct {
	region    string
	tableName string
}

type tenantItem struct {
	KeyId  string        `json:"id"`
	Config tenant.Config `json:"tenantConfig"`
}

func NewTenantStorer(config config.Config) (*TenantStorer, error) {
	region := config.GetString("ears.storage.tenant.region")
	if region == "" {
		return nil, &MissingConfigError{"ears.storage.tenant.region"}
	}
	tableName := config.GetString("ears.storage.tenant.tableName")
	if tableName == "" {
		return nil, &MissingConfigError{"ears.storage.tenant.tableName"}
	}

	return &TenantStorer{
		region:    region,
		tableName: tableName,
	}, nil
}

func (s *TenantStorer) GetAllConfigs(ctx context.Context) ([]tenant.Config, error) {

	ctx, span := db.CreateSpan(ctx, "getAllTenantConfigs", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(s.tableName))
	defer span.End()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.region),
	})
	if err != nil {
		return nil, &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	input := &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	}
	tenants := make([]tenant.Config, 0)
	for {
		result, err := svc.ScanWithContext(ctx, input)
		if err != nil {
			return nil, &DynamoDbGetItemError{err}
		}
		for _, item := range result.Items {
			var t tenantItem
			err = dynamodbattribute.UnmarshalMap(item, &t)
			if err != nil {
				return nil, &DynamoDbMarshalError{err}
			}
			tenants = append(tenants, t.Config)
		}
		if result.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = result.LastEvaluatedKey
	}
	return tenants, nil
}

func (s *TenantStorer) GetConfig(ctx context.Context, id tenant.Id) (*tenant.Config, error) {

	ctx, span := db.CreateSpan(ctx, "getTenantConfig", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(s.tableName))
	defer span.End()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.region),
	})
	if err != nil {
		return nil, &tenant.InternalStorageError{Wrapped: err}
	}
	svc := dynamodb.New(sess)
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id.Key()),
			},
		},
		TableName: aws.String(s.tableName),
	}

	result, err := svc.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, &tenant.InternalStorageError{Wrapped: err}
	}

	if result.Item == nil {
		return nil, &tenant.TenantNotFoundError{Tenant: id}
	}

	var item tenantItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, &tenant.InternalStorageError{Wrapped: err}
	}

	return &item.Config, nil
}

func (s *TenantStorer) SetConfig(ctx context.Context, config tenant.Config) error {

	ctx, span := db.CreateSpan(ctx, "setTenantConfig", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(s.tableName))
	defer span.End()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.region),
	})
	if err != nil {
		return &tenant.InternalStorageError{Wrapped: err}
	}
	svc := dynamodb.New(sess)

	config.Modified = time.Now().Unix()
	configItem := tenantItem{
		KeyId:  config.Tenant.Key(),
		Config: config,
	}

	item, err := dynamodbattribute.MarshalMap(configItem)
	if err != nil {
		return &tenant.BadConfigError{}
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(s.tableName),
	}

	_, err = svc.PutItemWithContext(ctx, input)
	if err != nil {
		return &tenant.InternalStorageError{Wrapped: err}
	}
	return nil
}

func (s *TenantStorer) DeleteConfig(ctx context.Context, id tenant.Id) error {

	ctx, span := db.CreateSpan(ctx, "deleteTenantConfig", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(s.tableName))
	defer span.End()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.region),
	})
	if err != nil {
		return &tenant.InternalStorageError{Wrapped: err}
	}
	svc := dynamodb.New(sess)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id.Key()),
			},
		},
		TableName: aws.String(s.tableName),
	}

	_, err = svc.DeleteItemWithContext(ctx, input)
	if err != nil {
		return &tenant.InternalStorageError{Wrapped: err}
	}

	return nil
}
