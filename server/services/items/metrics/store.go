package metrics

import (
	"sync/atomic"
	"time"
)

type StoreAggregator struct {
	readsSec  *uint32
	readsMin  *uint32
	readsHr   *uint32
	writesSec *uint32
	writesMin *uint32
	writesHr  *uint32
	length    *uint64
	size      *uint64
}

func NewStoreAggregator(length, size uint64) *StoreAggregator {
	return &StoreAggregator{
		readsSec:  new(uint32),
		readsMin:  new(uint32),
		readsHr:   new(uint32),
		writesSec: new(uint32),
		writesMin: new(uint32),
		writesHr:  new(uint32),
		length:    &length,
		size:      &size,
	}
}

func (a *StoreAggregator) IncReads() {
	atomic.AddUint32(a.readsSec, 1)
}

func (a *StoreAggregator) IncWrites(lengthDelta, sizeDelta uint64) {
	atomic.AddUint32(a.writesSec, 1)

	if lengthDelta != 0 {
		atomic.AddUint64(a.length, lengthDelta)
	}

	if sizeDelta != 0 {
		atomic.AddUint64(a.size, sizeDelta)
	}
}

func (a *StoreAggregator) OnTickSecond(tick time.Time) {
	if r := atomic.LoadUint32(a.readsSec); r != 0 {
		atomic.AddUint32(a.readsMin, r)
		atomic.StoreUint32(a.readsSec, 0)
	}

	if w := atomic.LoadUint32(a.writesSec); w != 0 {
		atomic.AddUint32(a.writesMin, w)
		atomic.StoreUint32(a.writesSec, 0)
	}
}

func (a *StoreAggregator) OnTickMinute(tick time.Time) {
	atomic.AddUint32(a.readsHr, atomic.LoadUint32(a.readsMin))
	atomic.StoreUint32(a.readsMin, 0)

	atomic.AddUint32(a.writesHr, atomic.LoadUint32(a.writesMin))
	atomic.StoreUint32(a.writesMin, 0)
}

func (a *StoreAggregator) OnTickHour(tick time.Time) {
	atomic.StoreUint32(a.readsHr, 0)
	atomic.StoreUint32(a.writesHr, 0)
}

type StoreSnapshot struct {
	// Open   uint32 `json:"open"`
	ReadsSec  uint32 `json:"readsSec"`
	ReadsMin  uint32 `json:"readsMin"`
	ReadsHr   uint32 `json:"readsHr"`
	WritesSec uint32 `json:"writesSec"`
	WritesMin uint32 `json:"writesMin"`
	WritesHr  uint32 `json:"writesHr"`
	Length    uint64 `json:"length"`
	Size      uint64 `json:"size"`
}

func (a *StoreAggregator) Collect() *StoreSnapshot {
	return &StoreSnapshot{
		ReadsSec:  atomic.LoadUint32(a.readsSec),
		ReadsMin:  atomic.LoadUint32(a.readsMin),
		ReadsHr:   atomic.LoadUint32(a.readsHr),
		WritesSec: atomic.LoadUint32(a.writesSec),
		WritesMin: atomic.LoadUint32(a.writesMin),
		WritesHr:  atomic.LoadUint32(a.writesHr),
		Length:    atomic.LoadUint64(a.length),
		Size:      atomic.LoadUint64(a.size),
	}
}
