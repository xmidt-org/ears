// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

type OptionError struct {
	Err error
}

type Option func(OptionProcessor) error

type OptionProcessor interface {
	WithName(name string) error
	WithVersion(version string) error
	WithCommitID(commitID string) error

	WithConfig(config interface{}) error

	WithNewPluginer(fn NewPluginerFn) error
	WithPluginerHasher(fn HashFn) error

	// TODO: https://github.com/xmidt-org/ears/issues/74
	WithNewFilterer(fn NewFiltererFn) error
	WithFilterHasher(fn HashFn) error

	WithNewReceiver(fn NewReceiverFn) error
	WithReceiverHasher(fn HashFn) error

	WithNewSender(fn NewSenderFn) error
	WithSenderHasher(fn HashFn) error
}

func WithName(name string) Option {
	return func(o OptionProcessor) error {
		return o.WithName(name)
	}
}

func WithVersion(version string) Option {
	return func(o OptionProcessor) error {
		return o.WithVersion(version)
	}
}

func WithCommitID(commitID string) Option {
	return func(o OptionProcessor) error {
		return o.WithCommitID(commitID)
	}
}

func WithConfig(config interface{}) Option {
	return func(o OptionProcessor) error {
		return o.WithConfig(config)
	}
}

func WithNewPluginer(fn NewPluginerFn) Option {
	return func(o OptionProcessor) error {
		return o.WithNewPluginer(fn)
	}
}

func WithPluginerHasher(fn HashFn) Option {
	return func(o OptionProcessor) error {
		return o.WithPluginerHasher(fn)
	}
}

// TODO: https://github.com/xmidt-org/ears/issues/74
func WithNewFilterer(fn NewFiltererFn) Option {
	return func(o OptionProcessor) error {
		return o.WithNewFilterer(fn)
	}
}

func WithFilterHasher(fn HashFn) Option {
	return func(o OptionProcessor) error {
		return o.WithFilterHasher(fn)
	}
}

func WithNewReceiver(fn NewReceiverFn) Option {
	return func(o OptionProcessor) error {
		return o.WithNewReceiver(fn)
	}
}

func WithReceiverHasher(fn HashFn) Option {
	return func(o OptionProcessor) error {
		return o.WithReceiverHasher(fn)
	}
}

func WithNewSender(fn NewSenderFn) Option {
	return func(o OptionProcessor) error {
		return o.WithNewSender(fn)
	}
}

func WithSenderHasher(fn HashFn) Option {
	return func(o OptionProcessor) error {
		return o.WithSenderHasher(fn)
	}
}
