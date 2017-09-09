package metrics

import "github.com/js13kgames/glitchd/server"

type Aggregator interface {
	Bootstrap(ticks *server.TickManager)
	Collect() interface{}
}
