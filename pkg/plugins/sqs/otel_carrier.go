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

type SqsMessageCarrier struct {
	msg *sqs.Message
}

func NewSqsMessageCarrier(msg *sqs.Message) *SqsMessageCarrier {
	return &SqsMessageCarrier{msg: msg}
}

func (c *SqsMessageCarrier) Get(key string) string {
	for attrKey, attrVal := range c.msg.Attributes {
		if attrKey == key {
			return *attrVal
		}
	}
	return ""
}

func (c *SqsMessageCarrier) Set(key, val string) {
	c.msg.Attributes[key] = aws.String(val)
}

func (c *SqsMessageCarrier) Keys() []string {
	out := make([]string, len(c.msg.Attributes))
	i := 0
	for key := range c.msg.Attributes {
		out[i] = key
		i++
	}
	return out
}

type SqsSendMessageBatchRequestEntryCarrier struct {
	msg *sqs.SendMessageBatchRequestEntry
}

func NewSqsSendMessageBatchRequestEntryCarrier(msg *sqs.SendMessageBatchRequestEntry) *SqsSendMessageBatchRequestEntryCarrier {
	return &SqsSendMessageBatchRequestEntryCarrier{msg: msg}
}

func (c *SqsSendMessageBatchRequestEntryCarrier) Get(key string) string {
	val, ok := c.msg.MessageAttributes[key]
	if !ok {
		return ""
	}
	if val.StringValue == nil {
		return ""
	}
	return *val.StringValue
}

func (c *SqsSendMessageBatchRequestEntryCarrier) Set(key, val string) {
	if c.msg.MessageAttributes == nil {
		c.msg.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
	}
	c.msg.MessageAttributes[key] = &sqs.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(val),
	}
}

func (c *SqsSendMessageBatchRequestEntryCarrier) Keys() []string {
	out := make([]string, 0)
	for key, val := range c.msg.MessageAttributes {
		if val.StringValue != nil {
			out = append(out, key)
		}
	}
	return out
}
