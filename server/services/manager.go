package services

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/js13kgames/glitchd/server"
)

//
//
//
type Manager struct {
	interfaces []server.Interface
	services   []Service

	logger *zap.Logger
	ticker *time.Ticker

	onEachSecond []func(time time.Time)
	onEachMinute []func(time time.Time)
	onEachHour   []func(time time.Time)
}

//
//
//
func NewServiceManager(logger *zap.Logger, interfaces []server.Interface, services []Service) *Manager {
	return &Manager{
		logger:     logger,
		interfaces: interfaces,
		services:   services,
	}
}

//
//
//
func (manager *Manager) Bootstrap() {
	manager.logger.Debug("Bootstrapping services")
	for _, service := range manager.services {
		service.Bootstrap(manager, manager.interfaces, manager.services)
	}
}

//
//
//
func (manager *Manager) Run() {
	for _, iface := range manager.interfaces {
		go func(iface server.Interface) {
			manager.logger.Debug("Starting interface",
				zap.String("interface", iface.GetKind()),
				zap.String("action", "start"),
			)
			iface.Start()
		}(iface)
	}

	for _, service := range manager.services {
		go func(service Service) {
			manager.logger.Debug("Starting service",
				zap.String("service", service.GetName()),
				zap.String("action", "start"),
			)
			service.Start()
		}(service)
	}

	manager.ticker = time.NewTicker(time.Second * 1)

	for {
		select {
		// Note: Single per-second ticker appeared more efficient than separate tickers
		// for each time period we tick at.
		// Note: The handlers run in sequence. This could lead to races if they don't complete
		// their work within the allocated time before the next tick. However, we're not
		// even nearly there in terms of per tick load.
		// @monitor
		case tick := <-manager.ticker.C:
			go func() {
				for _, handler := range manager.onEachSecond {
					handler(tick)
				}
			}()

			// Naive second 0 = new minute.
			if tick.Second() == 0 {
				go func() {
					for _, handler := range manager.onEachMinute {
						handler(tick)
					}
				}()
			}

			// Naive minute 0 = new hour.
			if tick.Minute() == 0 {
				go func() {
					for _, handler := range manager.onEachHour {
						handler(tick)
					}
				}()
			}
		}
	}
}

//
//
//
func (manager *Manager) Stop(gracePeriod *time.Duration) {
	var deadline time.Time

	if gracePeriod != nil {
		deadline = time.Now().Add(*gracePeriod)
	}

	if manager.ticker != nil {
		manager.logger.Debug("Stopping tickers")
		manager.ticker.Stop()
	}

	wg := sync.WaitGroup{}
	wg.Add(len(manager.interfaces) + len(manager.services))

	// @todo Deadlines for services.
	for _, service := range manager.services {
		go func(service Service) {
			manager.logger.Debug("Stopping service",
				zap.String("service", service.GetName()),
				zap.String("action", "stop"),
			)
			service.Stop(&deadline)
			wg.Done()
		}(service)
	}

	for _, iface := range manager.interfaces {
		go func(iface server.Interface) {
			manager.logger.Debug("Stopping interface",
				zap.String("interface", iface.GetKind()),
				zap.String("action", "stop"),
			)
			iface.Stop(&deadline)
			wg.Done()
		}(iface)
	}

	wg.Wait()
}

func (manager *Manager) OnTickHour(handler server.TickHandler) {
	manager.onEachHour = append(manager.onEachHour, handler)
}

func (manager *Manager) OnTickMinute(handler server.TickHandler) {
	manager.onEachMinute = append(manager.onEachMinute, handler)
}

func (manager *Manager) OnTickSecond(handler server.TickHandler) {
	manager.onEachSecond = append(manager.onEachSecond, handler)
}
