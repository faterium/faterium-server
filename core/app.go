package core

import (
	"log"
	"sync"

	iface "github.com/ipfs/interface-go-ipfs-core"
	kuboCore "github.com/ipfs/kubo/core"
)

type App struct {
	services []Service

	IpfsApi  iface.CoreAPI
	IpfsNode *kuboCore.IpfsNode
}

type Service struct {
	name       string
	launchFunc LaunchFunc
}

type LaunchFunc func(app *App) error

// Add a named service with launch function
func (app *App) AddService(name string, service LaunchFunc) {
	app.services = append(app.services, Service{
		name:       name,
		launchFunc: service,
	})
}

// Launch all services with sync.WaitGroup
func (app *App) LaunchServices() {
	var wg sync.WaitGroup
	for _, ser := range app.services {
		wg.Add(1)
		// Launch service
		go func(service Service) {
			defer wg.Done()
			log.Printf("launchin' #%s service", service.name)
			if err := service.launchFunc(app); err != nil {
				log.Fatal(err)
			}
		}(ser)
	}
	wg.Wait()
}
