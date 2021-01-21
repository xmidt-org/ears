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
	"github.com/xmidt-org/ears/internal/pkg/app"
	"github.com/xmidt-org/ears/pkg/route"
	"time"
)

type DynamoDbStorer struct {
	region    string
	tableName string
}

func NewDynamoDbStorer(config app.Config) (*DynamoDbStorer, error) {
	region := config.GetString("ears.db.region")
	if region == "" {
		return nil, &MissingConfigError{"ears.db.region"}
	}
	tableName := config.GetString("ears.db.tableName")
	if tableName == "" {
		return nil, &MissingConfigError{"ears.db.tableName"}
	}

	return &DynamoDbStorer{
		region:    region,
		tableName: tableName,
	}, nil
}

func (d *DynamoDbStorer) getRoute(ctx context.Context, id string, svc *dynamodb.DynamoDB) (*route.Config, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(d.tableName),
	}

	result, err := svc.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, &DynamoDbGetItemError{err}
	}

	if result.Item == nil {
		return nil, &route.RouteNotFoundError{id}
	}

	var routeConfig route.Config
	err = dynamodbattribute.UnmarshalMap(result.Item, &routeConfig)
	if err != nil {
		return nil, &DynamoDbMarshalError{err}
	}

	return &routeConfig, nil
}

func (d *DynamoDbStorer) GetRoute(ctx context.Context, id string) (route.Config, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	empty := route.Config{}
	if err != nil {
		return empty, &DynamoDbNewSessionError{err}
	}

	svc := dynamodb.New(sess)
	r, err := d.getRoute(ctx, id, svc)
	if err != nil {
		return empty, err
	}
	return *r, nil
}

func (d *DynamoDbStorer) GetAllRoutes(ctx context.Context) ([]route.Config, error) {
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

	result, err := svc.ScanWithContext(ctx, input)
	if err != nil {
		return nil, &DynamoDbGetItemError{err}
	}

	routes := make([]route.Config, 0)
	for _, item := range result.Items {
		var routeConfig route.Config
		err = dynamodbattribute.UnmarshalMap(item, &routeConfig)
		if err != nil {
			return nil, &DynamoDbMarshalError{err}
		}
		routes = append(routes, routeConfig)
	}
	return routes, nil
}

func (d *DynamoDbStorer) setRoute(ctx context.Context, r route.Config, svc *dynamodb.DynamoDB) error {
	//First see if the route already exists
	oldRoute, err := d.getRoute(ctx, r.Id, svc)
	if err != nil {
		return err
	}
	r.Created = time.Now().Unix()
	r.Modified = r.Created
	if oldRoute != nil {
		//Set created time from the old record
		r.Created = oldRoute.Created
	}

	item, err := dynamodbattribute.MarshalMap(r)
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
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	return d.setRoute(ctx, r, svc)
}

//TODO
func (d *DynamoDbStorer) SetRoutes(ctx context.Context, routes []route.Config) error {

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

func (d *DynamoDbStorer) deleteRoute(ctx context.Context, id string, svc *dynamodb.DynamoDB) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
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

func (d *DynamoDbStorer) DeleteRoute(ctx context.Context, id string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(d.region),
	})
	if err != nil {
		return &DynamoDbNewSessionError{err}
	}
	svc := dynamodb.New(sess)
	return d.deleteRoute(ctx, id, svc)
}

func (d *DynamoDbStorer) DeleteRoutes(ctx context.Context, ids []string) error {
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
		err = d.deleteRoute(ctx, id, svc)
		if err != nil {
			return err
		}
	}
	return nil
}
