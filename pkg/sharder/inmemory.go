package sharder

// trivial in memory node manager that always has one active node (itself)
type inmemoryNodeManager struct {
	identity string
}

func newInMemoryNodeManager(identity string, configData map[string]string) (*inmemoryNodeManager, error) {
	nodeManager := inmemoryNodeManager{
		identity: identity,
	}
	return &nodeManager, nil
}

func (d *inmemoryNodeManager) GetActiveNodes() ([]string, error) {
	return []string{"localhost"}, nil
}

func (d *inmemoryNodeManager) RemoveNode() {
}

func (d *inmemoryNodeManager) Stop() {
}
