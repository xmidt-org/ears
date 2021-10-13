package sharder

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"strings"
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
	server *dynamodb.DynamoDB
	tag    string
	region string
	// table name of the healthTable
	tableName string
	// frequency to read/write the state from/to dynamoDB
	frequency int
	// node state's last updated time is older than this (in seconds), treat as bad node
	// also if can't get/update node state for this time than it will exit the service
	older int
	// ip address of this node
	identity string
}

func newDynamoDBNodesManager(tableName, region, ip string, frq, older int, tag string) (*dynamoDBNodesManager, error) {
	d := &dynamoDBNodesManager{}
	d.region = region
	d.tableName = tableName
	if tag != "" {
		d.tag = tag
	} else {
		d.tag = "node"
	}
	d.identity = ip
	if frq > 0 {
		d.frequency = frq
	} else if frq == 0 {
		d.frequency = update_frequency
	} else {
		err := fmt.Errorf("update frequency value is invalid: %d", frq)
		return nil, err
	}
	if older > 0 {
		d.older = older
	} else if older == 0 {
		d.older = older_than
	} else {
		err := fmt.Errorf("old value is invalid: %d", older)
		return nil, err
	}
	// validate regions we can handle
	switch d.region {
	case "local":
		s, err := session.NewSession(&aws.Config{
			Endpoint: aws.String("http://127.0.0.1:8000")},
		)
		if nil != err {
			return nil, err
		}
		d.server = dynamodb.New(s)
	case "":
		return nil, errors.New("region must be specified")
	default:
		s, err := session.NewSession(&aws.Config{
			Region: aws.String(region)},
		)
		if nil != err {
			return nil, err
		}
		d.server = dynamodb.New(s)
	}
	// at this point we have a server
	err := d.ensureTable()
	if err != nil {
		return nil, err
	}
	go d.updateMyState()
	return d, nil
}

func (d *dynamoDBNodesManager) GetActiveNodes() ([]string, error) {
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
		if int(time.Since(updateTime).Seconds()) > d.older {
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

func (d *dynamoDBNodesManager) updateMyState() {
	defer d.RemoveNode()
	for {
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
		d.server.UpdateItem(input)
		time.Sleep(update_frequency * time.Second)
		continue
	}
}

// remove own record from health table
func (d *dynamoDBNodesManager) RemoveNode() {
	d.deleteRecord(d.identity)
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
