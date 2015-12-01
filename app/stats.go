package app

import (
	"sync/atomic"

	"github.com/ironsmile/nedomi/types"
)

type applicationStats types.AppStats

func (as *applicationStats) requested() uint64 {
	return atomic.AddUint64(&as.Requests, 1)
}
func (as *applicationStats) responded() uint64 {
	return atomic.AddUint64(&as.Responded, 1)
}

func (as *applicationStats) notConfigured() uint64 {
	return atomic.AddUint64(&as.NotConfigured, 1)
}
