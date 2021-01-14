package dynamo

type MissingConfigError struct {
	configName string
}

func (e *MissingConfigError) Error() string {
	return "MissingConfigError (configName=" + e.configName + ")"
}

type DynamoDbError struct {
	message string
	Source  error
}

func (e *DynamoDbError) Error() string {
	return "DynamoDbError (message=" + e.message + "): " + e.Source.Error()
}

func (e *DynamoDbError) Unwrap() error {
	return e.Source
}
