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
	"github.com/xmidt-org/ears/pkg/fragments"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/semconv/v1.4.0"
)

type DynamoDbFragmentStorer struct {
	region    string
	tableName string
}

type fragmentItem struct {
	KeyId    string             `json:"id"`
	TenantId tenant.Id          `json:"tenantId"`
	Config   route.PluginConfig `json:"pluginConfig"`
}

func NewDynamoDbFragmentStorer(config config.Config) (*DynamoDbFragmentStorer, error) {
	region := config.GetString("ears.storage.fragment.region")
	if region == "" {
		return nil, &MissingConfigError{"ears.storage.fragment.region"}
	}
	tableName := config.GetString("ears.storage.fragment.tableName")
	if tableName == "" {
		return nil, &MissingConfigError{"ears.storage.fragment.tableName"}
	}
	return &DynamoDbFragmentStorer{
		region:    region,
		tableName: tableName,
	}, nil
}

func (d *DynamoDbFragmentStorer) getFragment(ctx context.Context, tid tenant.Id, fragmentName string, svc *dynamodb.DynamoDB) (*route.PluginConfig, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(tid.KeyWithFragment(fragmentName)),
			},
		},
		TableName: aws.String(d.tableName),
	}
	result, err := svc.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, &DynamoDbGetItemError{err}
	}
	if result.Item == nil {
		return nil, &fragments.FragmentNotFoundError{TenantId: tid, FragmentName: fragmentName}
	}
	var item fragmentItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, &DynamoDbMarshalError{err}
	}
	fragmentConfig := item.Config
	return &fragmentConfig, nil
}

func (d *DynamoDbFragmentStorer) GetFragment(ctx context.Context, tid tenant.Id, id string) (route.PluginConfig, error) {
	ctx, span := db.CreateSpan(ctx, "getFragment", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	empty := route.PluginConfig{}
	if err != nil {
		return empty, &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	r, err := d.getFragment(ctx, tid, id, svc)
	if err != nil {
		return empty, err
	}
	return *r, nil
}

func (d *DynamoDbFragmentStorer) GetAllFragments(ctx context.Context) ([]route.PluginConfig, error) {
	ctx, span := db.CreateSpan(ctx, "getAllFragments", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return nil, &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	input := &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	}
	fragments := make([]route.PluginConfig, 0)
	for {
		result, err := svc.ScanWithContext(ctx, input)
		if err != nil {
			return nil, &DynamoDbGetItemError{err}
		}
		for _, item := range result.Items {
			var f fragmentItem
			err = dynamodbattribute.UnmarshalMap(item, &f)
			if err != nil {
				return nil, &DynamoDbMarshalError{err}
			}
			fragments = append(fragments, f.Config)
		}
		if result.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = result.LastEvaluatedKey
	}
	return fragments, nil
}

func (d *DynamoDbFragmentStorer) GetAllTenantFragments(ctx context.Context, tid tenant.Id) ([]route.PluginConfig, error) {
	ctx, span := db.CreateSpan(ctx, "getAllTenantFragments", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return nil, &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	input := &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	}
	fragments := make([]route.PluginConfig, 0)
	for {
		result, err := svc.ScanWithContext(ctx, input)
		if err != nil {
			return nil, &DynamoDbGetItemError{err}
		}
		for _, item := range result.Items {
			var f fragmentItem
			err = dynamodbattribute.UnmarshalMap(item, &f)
			if err != nil {
				return nil, &DynamoDbMarshalError{err}
			}
			if tid.Equal(f.TenantId) {
				fragments = append(fragments, f.Config)
			}
		}
		if result.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = result.LastEvaluatedKey
	}
	return fragments, nil
}

func (d *DynamoDbFragmentStorer) setFragment(ctx context.Context, tid tenant.Id, f route.PluginConfig, svc *dynamodb.DynamoDB) error {
	fragment := fragmentItem{
		KeyId:    tid.KeyWithFragment(f.Name),
		Config:   f,
		TenantId: tid,
	}
	av, err := dynamodbattribute.NewEncoder(func(e *dynamodbattribute.Encoder) {
		e.NullEmptyString = false
		e.NullEmptyByteSlice = false
		e.EnableEmptyCollections = true
	}).Encode(fragment)
	if err != nil {
		return &DynamoDbMarshalError{err}
	}
	input := &dynamodb.PutItemInput{
		Item:      av.M,
		TableName: aws.String(d.tableName),
	}
	_, err = svc.PutItemWithContext(ctx, input)
	if err != nil {
		return &DynamoDbPutItemError{err}
	}
	return nil
}

func (d *DynamoDbFragmentStorer) SetFragment(ctx context.Context, tid tenant.Id, f route.PluginConfig) error {
	ctx, span := db.CreateSpan(ctx, "storeFragment", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	return d.setFragment(ctx, tid, f, svc)
}

func (d *DynamoDbFragmentStorer) SetFragments(ctx context.Context, tid tenant.Id, fragments []route.PluginConfig) error {
	ctx, span := db.CreateSpan(ctx, "storeFragments", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	for _, f := range fragments {
		err = d.setFragment(ctx, tid, f, svc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DynamoDbFragmentStorer) deleteFragment(ctx context.Context, tid tenant.Id, id string, svc *dynamodb.DynamoDB) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(tid.KeyWithFragment(id)),
			},
		},
		TableName: aws.String(d.tableName),
	}
	_, err := svc.DeleteItemWithContext(ctx, input)
	if err != nil {
		return &DynamoDbDeleteItemError{err}
	}
	return nil
}

func (d *DynamoDbFragmentStorer) DeleteFragment(ctx context.Context, tid tenant.Id, id string) error {
	ctx, span := db.CreateSpan(ctx, "deleteFragment", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	return d.deleteFragment(ctx, tid, id, svc)
}

func (d *DynamoDbFragmentStorer) DeleteFragments(ctx context.Context, tid tenant.Id, ids []string) error {
	ctx, span := db.CreateSpan(ctx, "deleteFragments", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	for _, id := range ids {
		err = d.deleteFragment(ctx, tid, id, svc)
		if err != nil {
			return err
		}
	}
	return nil
}
