package dynamo

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/tenant"
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

func (s *TenantStorer) GetConfig(ctx context.Context, id tenant.Id) (*tenant.Config, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.region),
	})
	if err != nil {
		return nil, &tenant.InternalStorageError{err}
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
		return nil, &tenant.InternalStorageError{err}
	}

	if result.Item == nil {
		return nil, &tenant.TenantNotFoundError{id}
	}

	var item tenantItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, &tenant.InternalStorageError{err}
	}

	return &item.Config, nil
}

func (s *TenantStorer) SetConfig(ctx context.Context, config tenant.Config) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.region),
	})
	if err != nil {
		return &tenant.InternalStorageError{err}
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
		return &tenant.InternalStorageError{err}
	}
	return nil
}

func (s *TenantStorer) DeleteConfig(ctx context.Context, id tenant.Id) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.region),
	})
	if err != nil {
		return &tenant.InternalStorageError{err}
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
		return &tenant.InternalStorageError{err}
	}

	return nil
}
