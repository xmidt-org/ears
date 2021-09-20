package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SqsMessageAttributeCarrier struct {
	attributes map[string]*sqs.MessageAttributeValue
}

func NewSqsMessageAttributeCarrier(attributes map[string]*sqs.MessageAttributeValue) *SqsMessageAttributeCarrier {
	return &SqsMessageAttributeCarrier{attributes: attributes}
}

func (c *SqsMessageAttributeCarrier) Get(key string) string {
	val, ok := c.attributes[key]
	if !ok {
		return ""
	}
	if val.StringValue == nil {
		return ""
	}
	return *val.StringValue
}

func (c *SqsMessageAttributeCarrier) Set(key, val string) {
	c.attributes[key] = &sqs.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(val),
	}
}

func (c *SqsMessageAttributeCarrier) Keys() []string {
	out := make([]string, 0)
	for key, val := range c.attributes {
		if val.StringValue != nil {
			out = append(out, key)
		}
	}
	return out
}
