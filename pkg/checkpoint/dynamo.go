package checkpoint

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"strings"
	"time"
)

// dynamodb default values
const readCap = 10
const writeCap = 10
const keyName = "key"
const sortKeyName = "id"
const valueNameSequenceId = "sequenceId"
const valueNameLastUpdated = "lastUpdated"

func newDynamoCheckpointManager(config StorageConfig) (CheckpointManager, error) {
	cm := DynamoCheckpointManager{
		StorageConfig: config,
		lastUpdate:    make(map[string]time.Time),
	}
	switch cm.Region {
	case "local":
		session, err := session.NewSession(&aws.Config{
			Endpoint: aws.String("http://127.0.0.1:8000")},
		)
		if nil != err {
			return nil, err
		}
		cm.svc = dynamodb.New(session)
	case "":
		return nil, errors.New("region must be specified")
	default:
		session, err := session.NewSession(&aws.Config{
			Region: aws.String(cm.Region)},
		)
		if nil != err {
			return nil, err
		}
		cm.svc = dynamodb.New(session)
	}
	err := cm.ensureTable()
	if err != nil {
		return nil, err
	}
	return &cm, nil
}

func (cm *DynamoCheckpointManager) GetCheckpoint(Id string) (string, time.Time, error) {
	keyCond := expression.Key(keyName).Equal(expression.Value(cm.Env))
	proj := expression.NamesList(expression.Name(sortKeyName), expression.Name(valueNameSequenceId), expression.Name(valueNameLastUpdated))
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		WithProjection(proj).
		Build()
	if err != nil {
		return "", time.Time{}, err
	}
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(cm.Table),
	}
	results, err := cm.svc.Query(input)
	if err != nil {
		return "", time.Time{}, err
	}
	if len(results.Items) == 0 {
		return "", time.Time{}, nil
	} else if len(results.Items) > 1 {
		return "", time.Time{}, errors.New("found more than one checkpoint for Id " + Id)
	} else {
		sequenceIdAttr, found := results.Items[0][valueNameSequenceId]
		if !found {
			return "", time.Time{}, errors.New("invalid checkpoint for Id " + Id)
		}
		lastUpdatedAttr, found := results.Items[0][valueNameLastUpdated]
		if !found {
			return "", time.Time{}, errors.New("invalid checkpoint for Id " + Id)
		}
		updateTime, err := time.Parse(time.RFC3339Nano, aws.StringValue(lastUpdatedAttr.S))
		if err != nil {
			return "", time.Time{}, err
		}
		return aws.StringValue(sequenceIdAttr.S), updateTime, nil
	}
}

func (cm *DynamoCheckpointManager) timestamp(t time.Time) string {
	ts := t.UTC().Format(time.RFC3339Nano)
	// format library cuts off trailing 0s. See: https://github.com/golang/go/issues/19635
	// this causes problems for string based sorting
	if !strings.Contains(ts, ".") {
		ts = ts[0:len(ts)-1] + ".000000000Z"
	}
	return ts
}

func (cm *DynamoCheckpointManager) SetCheckpoint(Id string, sequenceNumber string) error {
	cm.Lock()
	lastUpdate, ok := cm.lastUpdate[Id]
	update := false
	if !ok {
		update = true
		cm.lastUpdate[Id] = time.Now()
	} else if ok && int(time.Since(lastUpdate).Seconds()) >= cm.UpdateFrequencySeconds {
		update = true
		cm.lastUpdate[Id] = time.Now()
	}
	cm.Unlock()
	if update {
		input := &dynamodb.UpdateItemInput{
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":u": {
					S: aws.String(sequenceNumber),
				},
				":v": {
					S: aws.String(cm.timestamp(time.Now())),
				},
			},
			TableName: aws.String(cm.Table),
			Key: map[string]*dynamodb.AttributeValue{
				keyName: {
					S: aws.String(cm.Env),
				},
				sortKeyName: {
					S: aws.String(Id),
				},
			},
			ReturnValues:     aws.String("UPDATED_NEW"),
			UpdateExpression: aws.String("set " + valueNameSequenceId + " = :u , " + valueNameLastUpdated + " = :v"),
		}
		_, err := cm.svc.UpdateItem(input)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cm *DynamoCheckpointManager) GetId(pluginHash, consumerName, streamName, shardID string) string {
	return cm.Env + "-" + pluginHash + "-" + consumerName + "-" + streamName + "-" + shardID
}

func (cm *DynamoCheckpointManager) ensureTable() error {
	describe := &dynamodb.DescribeTableInput{
		TableName: aws.String(cm.Table),
	}
	_, err := cm.svc.DescribeTable(describe)
	if err != nil {
		createTableInput := cm.tableDescription()
		_, err = cm.svc.CreateTable(createTableInput)
		if err != nil {
			return err
		}
		if err := cm.svc.WaitUntilTableExists(describe); nil != err {
			return err
		}
	}
	return nil
}

func (cm *DynamoCheckpointManager) tableDescription() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: aws.String(cm.Table),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(keyName),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(sortKeyName),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(keyName),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(sortKeyName),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(readCap),
			WriteCapacityUnits: aws.Int64(writeCap),
		},
	}
}
