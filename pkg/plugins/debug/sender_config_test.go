package debug_test

import (
	"testing"

	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xorcare/pointer"

	. "github.com/onsi/gomega"
)

func TestWithDefault(t *testing.T) {

	testCases := []struct {
		name     string
		input    debug.SenderConfig
		expected debug.SenderConfig
	}{
		{
			name:     "empty",
			input:    debug.SenderConfig{},
			expected: debug.DefaultSenderConfig,
		},
		{
			name: "maxHistory",
			input: debug.SenderConfig{
				MaxHistory: pointer.Int(12),
			},
			expected: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  pointer.Int(12),
				Writer:      debug.DefaultSenderConfig.Writer,
			},
		},

		{
			name: "destination",
			input: debug.SenderConfig{
				Destination: debug.DestinationStderr,
			},
			expected: debug.SenderConfig{
				Destination: debug.DestinationStderr,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
		},

		{
			name: "writer",
			input: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				Writer:      &debug.SendSlice{},
			},
			expected: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      &debug.SendSlice{},
			},
		},

		{
			name: "all",
			input: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  pointer.Int(19),
				Writer:      &debug.SendStdout{},
			},
			expected: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  pointer.Int(19),
				Writer:      &debug.SendStdout{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			d := tc.input.WithDefaults()
			a.Expect(d).To(Equal(tc.expected))

		})
	}

}

func TestValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input debug.SenderConfig
		valid bool
	}{
		{
			name:  "default",
			input: debug.DefaultSenderConfig,
			valid: true,
		},
		{
			name:  "with-defaults",
			input: (&debug.SenderConfig{}).WithDefaults(),
			valid: true,
		},

		{
			name:  "empty",
			input: debug.SenderConfig{},
			valid: false,
		},

		{
			name: "full",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: true,
		},

		{
			name: "destination-unset",
			input: debug.SenderConfig{
				MaxHistory: debug.DefaultSenderConfig.MaxHistory,
				Writer:     debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "destination-unknown",
			input: debug.SenderConfig{
				Destination: debug.DestinationUnknown,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "destination-23",
			input: debug.SenderConfig{
				Destination: 23,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "maxHistory-negative",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  pointer.Int(-1),
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "maxHistory-nil",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  nil,
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: false,
		},

		{
			name: "maxHistory-zero",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  pointer.Int(0),
				Writer:      debug.DefaultSenderConfig.Writer,
			},
			valid: true,
		},

		{
			name: "writer-nil",
			input: debug.SenderConfig{
				Destination: debug.DefaultSenderConfig.Destination,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      nil,
			},
			valid: true,
		},

		{
			name: "writer-custom",
			input: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      &debug.SendStderr{},
			},
			valid: true,
		},

		{
			name: "writer-custom-but-nil",
			input: debug.SenderConfig{
				Destination: debug.DestinationCustom,
				MaxHistory:  debug.DefaultSenderConfig.MaxHistory,
				Writer:      nil,
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewWithT(t)

			err := tc.input.Validate()
			if tc.valid {
				a.Expect(err).To(BeNil())
			} else {
				a.Expect(err).ToNot(BeNil())
			}

		})

	}

}
