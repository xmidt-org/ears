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
const defaultUpdateFrequency = 10
const defaultOlderThan = 60
const readCap = 10
const writeCap = 10

const keyName = "key"
const sortKeyName = "ip"
const valueName = "lastUpdated"

//dynamoDB as the node states manager
type dynamoDBNodeManager struct {
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
	// cached nodes
	activeNodes []string
	// last update time
	lastUpdateTime time.Time
}

func newDynamoDBNodeManager(identity string, configData map[string]string) (*dynamoDBNodeManager, error) {
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
	nodeManager := &dynamoDBNodeManager{
		logger:          event.GetEventLogger(),
		region:          configData["region"],
		tableName:       tableName,
		tag:             configData["tag"],
		identity:        identity,
		updateFrequency: updateFrequency,
		olderThan:       olderThan,
		stopped:         false,
		activeNodes:     make([]string, 0),
		lastUpdateTime:  time.Now(),
	}
	if nodeManager.tag == "" {
		nodeManager.tag = "node"
	}
	if nodeManager.updateFrequency <= 0 {
		nodeManager.updateFrequency = defaultUpdateFrequency
	}
	if nodeManager.olderThan <= 0 {
		nodeManager.olderThan = defaultOlderThan
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
	nodeManager.logger.Info().Str("op", "sharder.newDynamoDBNodeManager").Str("identity", nodeManager.identity).Msg("starting node state manager")
	return nodeManager, nil
}

func (d *dynamoDBNodeManager) GetActiveNodes() ([]string, error) {
	if int(time.Since(d.lastUpdateTime).Seconds()) < d.updateFrequency && len(d.activeNodes) > 0 {
		return d.activeNodes, nil
	}
	activeNodes := make([]string, 0)
	keyCond := expression.Key(keyName).Equal(expression.Value(d.tag))
	proj := expression.NamesList(expression.Name(sortKeyName), expression.Name(valueName))
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
		ipAttr, found := record[sortKeyName]
		if !found { // ip address not existing in this record
			continue
		}
		lastUpdatedAttr, found := record[valueName]
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
	d.activeNodes = activeNodes
	d.lastUpdateTime = time.Now()
	return activeNodes, nil
}

func (d *dynamoDBNodeManager) timestamp(t time.Time) string {
	ts := t.UTC().Format(time.RFC3339Nano)
	// format library cuts off trailing 0s. See: https://github.com/golang/go/issues/19635
	// this causes problems for string based sorting
	if !strings.Contains(ts, ".") {
		ts = ts[0:len(ts)-1] + ".000000000Z"
	}
	return ts
}

func (d *dynamoDBNodeManager) updateState() {
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
					keyName: {
						S: aws.String(d.tag),
					},
					sortKeyName: {
						S: aws.String(d.identity),
					},
				},
				ReturnValues:     aws.String("UPDATED_NEW"),
				UpdateExpression: aws.String("set " + valueName + " = :u"),
			}
			_, err := d.server.UpdateItem(input)
			if err != nil {
				d.logger.Error().Str("op", "sharder.updateState").Str("identity", d.identity).Msg(err.Error())
			}
			time.Sleep(time.Duration(d.updateFrequency) * time.Second)
		}
	}()
}

// remove own record from health table
func (d *dynamoDBNodeManager) RemoveNode() {
	d.deleteRecord(d.identity)
}

func (d *dynamoDBNodeManager) Stop() {
	d.Lock()
	d.stopped = true
	d.RemoveNode()
	d.Unlock()
}

// delete stale record, not care if success of not
func (d *dynamoDBNodeManager) deleteRecord(ip string) {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			keyName: {
				S: aws.String(d.tag),
			},
			sortKeyName: {
				S: aws.String(ip),
			},
		},
		TableName: aws.String(d.tableName),
	}
	d.server.DeleteItem(input)
}

// health_table description
func (d *dynamoDBNodeManager) tableDescription() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: aws.String(d.tableName),
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

// ensureTable verifies the table already exists, otherwise creates it
func (d *dynamoDBNodeManager) ensureTable() error {
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
