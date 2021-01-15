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
