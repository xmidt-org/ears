package block

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
)

type Filter struct {
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
