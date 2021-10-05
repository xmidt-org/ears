package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"testing"
)

func TestSqsMessageAttributeCarrier(t *testing.T) {
	attr := make(map[string]*sqs.MessageAttributeValue)
	c := NewSqsMessageAttributeCarrier(attr)

	val := c.Get("key1")
	if val != "" {
		t.Fatalf("Expect empty attribute for key1, got %s\n", val)
	}

	c.Set("key1", "val1")
	val = c.Get("key1")
	if val != "val1" {
		t.Fatalf("Expect val1 for key1, got %s\n", val)
	}

	c.Set("key2", "val2")
	val = c.Get("key2")
	if val != "val2" {
		t.Fatalf("Expect val2 for key2, got %s\n", val)
	}

	c.Set("key1", "val3")
	val = c.Get("key1")
	if val != "val3" {
		t.Fatalf("Expect val3 for key1, got %s\n", val)
	}

	keys := c.Keys()
	if len(keys) != 2 {
		t.Fatalf("Expect 2 keys, got %d instead\n", len(keys))
	}
}
