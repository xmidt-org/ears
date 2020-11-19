package chain

import (
	"sync"

	"github.com/xmidt-org/ears/pkg/filter"
)

//go:generate rm -f testing_mock.go
//go:generate moq -out testing_mock.go . FiltererChain

// FiltererChain
type FiltererChain interface {
	filter.Filterer

	// Add will add a filterer to the chain
	Add(f filter.Filterer)
}

type Chain struct {
	sync.RWMutex

	filterers []filter.Filterer
}

type InvalidArgumentError struct {
	Err error
}
