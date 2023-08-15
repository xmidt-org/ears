package block

import (
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
)

type Filter struct {
	sync.RWMutex
	name                          string
	plugin                        string
	tid                           tenant.Id
	successCounter                int
	errorCounter                  int
	filterCounter                 int
	successVelocityCounter        int
	errorVelocityCounter          int
	filterVelocityCounter         int
	currentSuccessVelocityCounter int
	currentErrorVelocityCounter   int
	currentFilterVelocityCounter  int
	currentSec                    int64
}
