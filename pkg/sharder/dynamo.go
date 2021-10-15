package sharder

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/panics"
	"strconv"
	"strings"
	"sync"
	"time"
)

// dynamodb default values
const update_frequency = 10
const older_than = 60
const read_cap = 10
const write_cap = 10

const key_name = "key"
const sort_key_name = "ip"
const value_name = "lastUpdated"

//dynamoDB as the node states manager
type dynamoDBNodesManager struct {
	sync.Mutex
	server *dynamodb.DynamoDB
	tag    string
	region string
	// table name of the healthTable
	tableName string
	// updateFrequency to read/write the state from/to dynamoDB
	updateFrequency int
	// node state's last updated time is olderThan than this (in seconds), treat as bad node
	// also if can't get/update node state for this time than it will exit the service
	olderThan int
	// ip address of this node
	identity string
	// logger
	logger *zerolog.Logger
	//
	stopped bool
}

func newDynamoDBNodesManager(identity string, configData map[string]string) (*dynamoDBNodesManager, error) {
	tableName, found := configData["table"]
	if !found {
		return nil, errors.New("table name must be set")
	}
	var updateFrequency, olderThan int
	if frq, found := configData["updateFrequency"]; found {
		updateFrequency, _ = strconv.Atoi(frq)
	}
	if old, found := configData["olderThan"]; found {
		olderThan, _ = strconv.Atoi(old)
	}
	nodeManager := &dynamoDBNodesManager{
		logger:          event.GetEventLogger(),
		region:          configData["region"],
		tableName:       tableName,
		tag:             configData["tag"],
		identity:        identity,
		updateFrequency: updateFrequency,
		olderThan:       olderThan,
		stopped:         false,
	}
	if nodeManager.tag == "" {
		nodeManager.tag = "node"
	}
	if nodeManager.updateFrequency <= 0 {
		nodeManager.updateFrequency = update_frequency
	}
	if nodeManager.olderThan <= 0 {
		nodeManager.olderThan = older_than
	}
	// validate regions we can handle
	switch nodeManager.region {
	case "local":
		session, err := session.NewSession(&aws.Config{
			Endpoint: aws.String("http://127.0.0.1:8000")},
		)
		if nil != err {
			return nil, err
		}
		nodeManager.server = dynamodb.New(session)
	case "":
		return nil, errors.New("region must be specified")
	default:
		session, err := session.NewSession(&aws.Config{
			Region: aws.String(nodeManager.region)},
		)
		if nil != err {
			return nil, err
		}
		nodeManager.server = dynamodb.New(session)
	}
	err := nodeManager.ensureTable()
	if err != nil {
		return nil, err
	}
	nodeManager.updateState()
	nodeManager.logger.Info().Str("op", "sharder.newDynamoDBNodesManager").Str("identity", nodeManager.identity).Msg("starting node state manager")
	return nodeManager, nil
}

func (d *dynamoDBNodesManager) GetActiveNodes() ([]string, error) {
	//TODO: cache result for short period of time
	activeNodes := make([]string, 0)
	keyCond := expression.Key(key_name).Equal(expression.Value(d.tag))
	proj := expression.NamesList(expression.Name(sort_key_name), expression.Name(value_name))
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		WithProjection(proj).
		Build()
	if err != nil {
		return nil, err
	}
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(d.tableName),
	}
	results, err := d.server.Query(input)
	if err != nil {
		return nil, err
	}
	for _, record := range results.Items {
		ipAttr, found := record[sort_key_name]
		if !found { // ip address not existing in this record
			continue
		}
		lastUpdatedAttr, found := record[value_name]
		if !found {
			continue
		}
		lastUpdated := aws.StringValue(lastUpdatedAttr.S)
		updateTime, err := time.Parse(time.RFC3339Nano, lastUpdated)
		if err != nil {
			return nil, err
		}
		if int(time.Since(updateTime).Seconds()) > d.olderThan {
			d.deleteRecord(aws.StringValue(ipAttr.S))
		}
		activeNodes = append(activeNodes, aws.StringValue(ipAttr.S))
	}
	return activeNodes, nil
}

func (d *dynamoDBNodesManager) timestamp(t time.Time) string {
	ts := t.UTC().Format(time.RFC3339Nano)
	// format library cuts off trailing 0s. See: https://github.com/golang/go/issues/19635
	// this causes problems for string based sorting
	if !strings.Contains(ts, ".") {
		ts = ts[0:len(ts)-1] + ".000000000Z"
	}
	return ts
}

func (d *dynamoDBNodesManager) updateState() {
	go func() {
		defer func() {
			p := recover()
			if p != nil {
				panicErr := panics.ToError(p)
				d.logger.Error().Str("op", "sharder.updateMyState").Str("error", panicErr.Error()).
					Str("stackTrace", panicErr.StackTrace()).Msg("a panic has occurred in state updater")
			}
		}()
		for {
			d.Lock()
			s := d.stopped
			d.Unlock()
			if s {
				d.logger.Info().Str("op", "sharder.updateState").Str("identity", d.identity).Msg("stopping node state manager")
				defaultNodeStateManager = nil
				return
			}
			input := &dynamodb.UpdateItemInput{
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":u": {
						S: aws.String(d.timestamp(time.Now())),
					},
				},
				TableName: aws.String(d.tableName),
				Key: map[string]*dynamodb.AttributeValue{
					key_name: {
						S: aws.String(d.tag),
					},
					sort_key_name: {
						S: aws.String(d.identity),
					},
				},
				ReturnValues:     aws.String("UPDATED_NEW"),
				UpdateExpression: aws.String("set " + value_name + " = :u"),
			}
			_, err := d.server.UpdateItem(input)
			if err != nil {
				d.logger.Error().Str("op", "sharder.updateState").Str("identity", d.identity).Msg(err.Error())
			}
			time.Sleep(update_frequency * time.Second)
		}
	}()
}

// remove own record from health table
func (d *dynamoDBNodesManager) RemoveNode() {
	d.deleteRecord(d.identity)
}

func (d *dynamoDBNodesManager) Stop() {
	d.Lock()
	d.stopped = true
	d.RemoveNode()
	d.Unlock()
}

// delete stale record, not care if success of not
func (d *dynamoDBNodesManager) deleteRecord(ip string) {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			key_name: {
				S: aws.String(d.tag),
			},
			sort_key_name: {
				S: aws.String(ip),
			},
		},
		TableName: aws.String(d.tableName),
	}
	d.server.DeleteItem(input)
}

// health_table description
func (d *dynamoDBNodesManager) tableDescription() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: aws.String(d.tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(key_name),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(sort_key_name),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(key_name),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(sort_key_name),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(read_cap),
			WriteCapacityUnits: aws.Int64(write_cap),
		},
	}
}

// ensureTable verifies the table already exists, otherwise creates it
func (d *dynamoDBNodesManager) ensureTable() error {
	describe := &dynamodb.DescribeTableInput{
		TableName: aws.String(d.tableName),
	}
	_, err := d.server.DescribeTable(describe)
	if err != nil {
		createTableInput := d.tableDescription()
		_, err = d.server.CreateTable(createTableInput)
		if err != nil {
			return err
		}
		if err := d.server.WaitUntilTableExists(describe); nil != err {
			return err
		}
	}
	return nil
}
