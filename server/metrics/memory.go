package metrics

import "runtime"

type MemorySnapshot struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"totalAlloc"`
	Sys          uint64 `json:"sys"`
	Lookups      uint64 `json:"lookups"`
	Mallocs      uint64 `json:"mallocs"`
	Frees        uint64 `json:"frees"`
	HeapAlloc    uint64 `json:"heapAlloc"`
	HeapSys      uint64 `json:"heapSys"`
	HeapIdle     uint64 `json:"heapIdle"`
	HeapInUse    uint64 `json:"heapInUse"`
	HeapReleased uint64 `json:"heapReleased"`
	HeapObjects  uint64 `json:"heapObjects"`
	PauseTotalNs uint64 `json:"pauseTotal"`
	NumGC        uint32 `json:"numGC"`
	NumGoroutine int    `json:"numGoroutines"`
}

func newMemorySnapshot() *MemorySnapshot {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return &MemorySnapshot{
		Alloc:        mem.Alloc,
		TotalAlloc:   mem.TotalAlloc,
		Sys:          mem.Sys,
		Lookups:      mem.Lookups,
		Mallocs:      mem.Mallocs,
		Frees:        mem.Frees,
		HeapAlloc:    mem.HeapAlloc,
		HeapSys:      mem.HeapSys,
		HeapIdle:     mem.HeapIdle,
		HeapInUse:    mem.HeapInuse,
		HeapReleased: mem.HeapReleased,
		PauseTotalNs: mem.PauseTotalNs,
		NumGC:        mem.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
	}
}
