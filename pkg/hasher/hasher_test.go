package hasher_test

import (
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/hasher"
)

func TestString(t *testing.T) {
	testCases := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "alpha", value: "abcdefghijklmnopqrstuvwxyz"},
		{name: "numeric", value: "1234567890"},
		{name: "bytes", value: string([]byte{1, 2, 3, 4})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))

			g.Assert(t, tc.name, []byte(hasher.String(tc.value)))
		})
	}
}

func TestHash(t *testing.T) {
	testCases := []struct {
		name  string
		value interface{}
	}{
		{name: "string", value: ""},
		{name: "int", value: "abcdefghijklmnopqrstuvwxyz"},
		{name: "float", value: "1234567890"},
		{name: "bytes", value: []byte{1, 2, 3, 4}},
		{name: "[]string]", value: []string{"1234567890", "0987654321"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))

			g.Assert(t, tc.name, []byte(hasher.String(tc.value)))
		})
	}
}
