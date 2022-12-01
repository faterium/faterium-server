package core

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	iface "github.com/ipfs/interface-go-ipfs-core"
	kuboCore "github.com/ipfs/kubo/core"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
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

func TryImportCollections(pbApp *pocketbase.PocketBase, path string) {
	jsonFile, err := os.ReadFile(path)
	if err != nil {
		log.Print("collections.json not found, proceeding without import")
	} else {
		log.Print("collections.json found! importing...")

		var collections []*models.Collection
		err = json.Unmarshal(jsonFile, &collections)
		if err != nil {
			log.Print("collections.json is invalid, proceeding without import")
		} else {
			pbApp.Dao().ImportCollections(collections, false, nil)
			log.Print("collections.json successfully imported!")
		}
	}
}
