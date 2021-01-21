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

package dynamo

type MissingConfigError struct {
	configName string
}

func (e *MissingConfigError) Error() string {
	return "MissingConfigError (configName=" + e.configName + ")"
}

type DynamoDbGetItemError struct {
	Source error
}

func (e *DynamoDbGetItemError) Error() string {
	return "DynamoDbGetItemError: " + e.Source.Error()
}

func (e *DynamoDbGetItemError) Unwrap() error {
	return e.Source
}

type DynamoDbMarshalError struct {
	Source error
}

func (e *DynamoDbMarshalError) Error() string {
	return "DynamoDbMarshalError: " + e.Source.Error()
}

func (e *DynamoDbMarshalError) Unwrap() error {
	return e.Source
}

type DynamoDbNewSessionError struct {
	Source error
}

func (e *DynamoDbNewSessionError) Error() string {
	return "DynamoDbNewSessionError: " + e.Source.Error()
}

func (e *DynamoDbNewSessionError) Unwrap() error {
	return e.Source
}

type DynamoDbPutItemError struct {
	Source error
}

func (e *DynamoDbPutItemError) Error() string {
	return "DynamoDbPutItemError: " + e.Source.Error()
}

func (e *DynamoDbPutItemError) Unwrap() error {
	return e.Source
}

type DynamoDbDeleteItemError struct {
	Source error
}

func (e *DynamoDbDeleteItemError) Error() string {
	return "DynamoDbDeleteItemError: " + e.Source.Error()
}

func (e *DynamoDbDeleteItemError) Unwrap() error {
	return e.Source
}
