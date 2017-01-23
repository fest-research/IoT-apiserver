package handler

import (
	"github.com/fest-research/iot-addon/pkg/apiserver/controller"
	"github.com/fest-research/iot-addon/pkg/apiserver/proxy"
)

type IServiceFactory interface {
	GetRegisteredServices() []IService
}

type ServiceFactory struct {
	proxy    *proxy.Proxy
	services []IService
}

func NewServiceFactory(proxy *proxy.Proxy) *ServiceFactory {
	factory := &ServiceFactory{proxy: proxy, services: make([]IService, 0)}
	factory.init()

	return factory
}

func (this *ServiceFactory) registerService(service IService) {
	this.services = append(this.services, service)
}

func (this *ServiceFactory) init() {
	// Version service
	this.registerService(NewVersionService(this.proxy.RawProxy))

	// Node service
	this.registerService(NewNodeService(this.proxy.ServerProxy, controller.NewNodeController()))

	// Pod service
	this.registerService(NewPodService(this.proxy.ServerProxy, controller.NewPodController()))

	// Event service
	this.registerService(NewEventService(this.proxy.RawProxy))

	// Kubernetes service
	this.registerService(NewKubeService(this.proxy.RawProxy))
}

func (this *ServiceFactory) GetRegisteredServices() []IService {
	return this.services
}
