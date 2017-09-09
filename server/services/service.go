package services

import "github.com/js13kgames/glitchd/server"

type Service interface {
	server.Runnable
	//
	GetName() string
	//
	Bootstrap(manager *Manager, interfaces []server.Interface, services []Service)
}
