package server

import "time"

type TickHandler func(tick time.Time)

type TickManager interface {
	OnTickSecond(handler TickHandler)
	OnTickMinute(handler TickHandler)
	OnTickHour(handler TickHandler)
}
