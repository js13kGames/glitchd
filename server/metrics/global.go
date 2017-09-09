package metrics

import (
	"os"
	"time"

	"github.com/js13kgames/glitchd/server"
)

type GlobalAggregator struct {
	startTime time.Time
	pid       int
	runtime   RuntimeInfo
	hostname  string
	requests  RequestsAggregator
}

func NewGlobalAggregator() *GlobalAggregator {
	hostname, _ := os.Hostname()

	return &GlobalAggregator{
		pid:       os.Getpid(),
		hostname:  hostname,
		runtime:   newRuntimeInfo(),
		startTime: time.Now(),
		requests:  *NewRequestsAggregator(),
	}
}

func (a *GlobalAggregator) Bootstrap(ticks server.TickManager) {
	a.requests.Bootstrap(ticks)
}

func (a *GlobalAggregator) BeginRequest() {
	a.requests.Begin()
}

func (a *GlobalAggregator) EndRequest() {
	a.requests.End()
}

type GlobalSnapshot struct {
	Pid      int               `json:"pid"`
	Version  string            `json:"version"`
	Branch   string            `json:"branch"`
	Commit   string            `json:"commit"`
	Hostname string            `json:"hostname"`
	Runtime  RuntimeInfo       `json:"runtime"`
	Memory   *MemorySnapshot   `json:"memory"`
	TimeNow  time.Time         `json:"now"`
	TimeUp   float64           `json:"uptime"`
	Requests *RequestsSnapshot `json:"requests"`
}

func (a *GlobalAggregator) Collect() interface{} {
	now := time.Now()

	return &GlobalSnapshot{
		Pid:      a.pid,
		Version:  server.Version,
		Branch:   server.Branch,
		Commit:   server.Commit,
		Hostname: a.hostname,
		Runtime:  a.runtime,
		Memory:   newMemorySnapshot(),
		TimeNow:  now,
		TimeUp:   now.Sub(a.startTime).Seconds(),
		Requests: a.requests.Collect(),
	}
}
