package metrics

import (
	"sync/atomic"
	"time"

	"github.com/js13kgames/glitchd/server"
)

type RequestsAggregator struct {
	// open   *uint32
	second *uint32
	minute *uint32
	hour   *uint32
	total  *uint32
}

func NewRequestsAggregator() *RequestsAggregator {
	return &RequestsAggregator{
		// open:   new(uint32),
		second: new(uint32),
		minute: new(uint32),
		hour:   new(uint32),
		total:  new(uint32),
	}
}

func (a *RequestsAggregator) Bootstrap(ticks server.TickManager) {
	ticks.OnTickSecond(a.onTickSecond)
	ticks.OnTickMinute(a.onTickMinute)
	ticks.OnTickHour(a.onTickHour)
}

func (a *RequestsAggregator) Begin() {
	// atomic.AddUint32(a.open, 1)
	atomic.AddUint32(a.second, 1)
	atomic.AddUint32(a.minute, 1)
	atomic.AddUint32(a.hour, 1)
	atomic.AddUint32(a.total, 1)
}

func (a *RequestsAggregator) End() {
	// atomic.AddUint32(a.open, ^uint32(0))
}

func (a *RequestsAggregator) onTickSecond(tick time.Time) {
	atomic.StoreUint32(a.second, 0)
}

func (a *RequestsAggregator) onTickMinute(tick time.Time) {
	atomic.StoreUint32(a.minute, 0)
}

func (a *RequestsAggregator) onTickHour(tick time.Time) {
	atomic.StoreUint32(a.hour, 0)
}

type RequestsSnapshot struct {
	// Open   uint32 `json:"open"`
	Second uint32 `json:"second"`
	Minute uint32 `json:"minute"`
	Hour   uint32 `json:"hour"`
	Total  uint32 `json:"total"`
}

func (a *RequestsAggregator) Collect() *RequestsSnapshot {
	return &RequestsSnapshot{
		// Open:   atomic.LoadUint32(a.open),
		Second: atomic.LoadUint32(a.second),
		Minute: atomic.LoadUint32(a.minute),
		Hour:   atomic.LoadUint32(a.hour),
		Total:  atomic.LoadUint32(a.total),
	}
}
