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
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/route"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/semconv"
	"time"
)

type DynamoDbStorer struct {
	region    string
	tableName string
}

type routeItem struct {
	KeyId  string       `json:"id"`
	Config route.Config `json:"routeConfig"`
}

func NewDynamoDbStorer(config config.Config) (*DynamoDbStorer, error) {
	region := config.GetString("ears.storage.route.region")
	if region == "" {
		return nil, &MissingConfigError{"ears.storage.route.region"}
	}
	tableName := config.GetString("ears.storage.route.tableName")
	if tableName == "" {
		return nil, &MissingConfigError{"ears.storage.route.tableName"}
	}
	return &DynamoDbStorer{
		region:    region,
		tableName: tableName,
	}, nil
}

func (d *DynamoDbStorer) getRoute(ctx context.Context, tid tenant.Id, routeId string, svc *dynamodb.DynamoDB) (*route.Config, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(tid.KeyWithRoute(routeId)),
			},
		},
		TableName: aws.String(d.tableName),
	}
	result, err := svc.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, &DynamoDbGetItemError{err}
	}
	if result.Item == nil {
		return nil, &route.RouteNotFoundError{TenantId: tid, RouteId: routeId}
	}
	var item routeItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, &DynamoDbMarshalError{err}
	}
	routeConfig := item.Config
	return &routeConfig, nil
}

func (d *DynamoDbStorer) GetRoute(ctx context.Context, tid tenant.Id, id string) (route.Config, error) {
	ctx, span := db.CreateSpan(ctx, "getRoute", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	empty := route.Config{}
	if err != nil {
		return empty, &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	r, err := d.getRoute(ctx, tid, id, svc)
	if err != nil {
		return empty, err
	}
	return *r, nil
}

func (d *DynamoDbStorer) GetAllRoutes(ctx context.Context) ([]route.Config, error) {

	ctx, span := db.CreateSpan(ctx, "getRoutes", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
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
	routes := make([]route.Config, 0)
	for {
		result, err := svc.ScanWithContext(ctx, input)
		if err != nil {
			return nil, &DynamoDbGetItemError{err}
		}
		for _, item := range result.Items {
			var r routeItem
			err = dynamodbattribute.UnmarshalMap(item, &r)
			if err != nil {
				return nil, &DynamoDbMarshalError{err}
			}
			routes = append(routes, r.Config)
		}
		if result.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = result.LastEvaluatedKey
	}
	return routes, nil
}

func (d *DynamoDbStorer) GetAllTenantRoutes(ctx context.Context, tid tenant.Id) ([]route.Config, error) {

	ctx, span := db.CreateSpan(ctx, "getTenantRoutes", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()

	routes, err := d.GetAllRoutes(ctx)
	if err != nil {
		return nil, err
	}
	filterRoutes := make([]route.Config, 0)
	for _, route := range routes {
		if route.TenantId.Equal(tid) {
			filterRoutes = append(filterRoutes, route)
		}
	}
	return filterRoutes, nil
}

func (d *DynamoDbStorer) setRoute(ctx context.Context, r route.Config, svc *dynamodb.DynamoDB) error {
	//First see if the route already exists
	oldRoute, err := d.getRoute(ctx, r.TenantId, r.Id, svc)
	if err != nil {
		var notFoundErr *route.RouteNotFoundError
		if !errors.As(err, &notFoundErr) {
			return err
		}
	}
	r.Created = time.Now().Unix()
	r.Modified = r.Created
	if oldRoute != nil {
		//Set created time from the old record
		r.Created = oldRoute.Created
	}
	route := routeItem{
		KeyId:  r.TenantId.KeyWithRoute(r.Id),
		Config: r,
	}
	item, err := dynamodbattribute.MarshalMap(route)
	if err != nil {
		return &DynamoDbMarshalError{err}
	}
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(d.tableName),
	}
	_, err = svc.PutItemWithContext(ctx, input)
	if err != nil {
		return &DynamoDbPutItemError{err}
	}
	return nil
}

func (d *DynamoDbStorer) SetRoute(ctx context.Context, r route.Config) error {

	ctx, span := db.CreateSpan(ctx, "storeRoute", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	return d.setRoute(ctx, r, svc)
}

//TODO: make this more efficient
func (d *DynamoDbStorer) SetRoutes(ctx context.Context, routes []route.Config) error {

	ctx, span := db.CreateSpan(ctx, "storeRoutes", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	//The following may be optimized with BatchWriteItem. Something to consider
	//in the future
	for _, r := range routes {
		err = d.setRoute(ctx, r, svc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DynamoDbStorer) deleteRoute(ctx context.Context, tid tenant.Id, id string, svc *dynamodb.DynamoDB) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(tid.KeyWithRoute(id)),
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

func (d *DynamoDbStorer) DeleteRoute(ctx context.Context, tid tenant.Id, id string) error {

	ctx, span := db.CreateSpan(ctx, "deleteRoute", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	return d.deleteRoute(ctx, tid, id, svc)
}

func (d *DynamoDbStorer) DeleteRoutes(ctx context.Context, tid tenant.Id, ids []string) error {
	ctx, span := db.CreateSpan(ctx, "deleteRoutes", semconv.DBSystemDynamoDB, rtsemconv.DBTable.String(d.tableName))
	defer span.End()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	//The following may be optimized with BatchWriteItem. Something to consider
	//in the future
	for _, id := range ids {
		err = d.deleteRoute(ctx, tid, id, svc)
		if err != nil {
			return err
		}
	}
	return nil
}
