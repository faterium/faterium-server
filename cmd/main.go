package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbCore "github.com/pocketbase/pocketbase/core"

	fateCore "github.com/faterium/faterium-server/core"

	"github.com/aofei/mimesniffer"
)

func main() {
	log.Print("let's get runnin' this server!")

	// Server App struct to controll services
	app := fateCore.App{}
	// Add IPFS node service
	app.AddService("ipfs", launchIpfs)
	// Add PocketBase service
	app.AddService("pocketbase", launchPocketBase)
	// Launch all services
	app.LaunchServices()
}

func launchPocketBase(app *fateCore.App) error {
	pbApp := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDebug:   false,
		DefaultDataDir: "./data/pb",
	})
	pbApp.OnBeforeServe().Add(func(e *pbCore.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/ipfs/:cid",
			Handler: func(c echo.Context) error {
				cid := c.PathParam("cid")
				file, err := fateCore.GetFileFromIpfs(c.Request().Context(), app, cid)
				if err != nil {
					return err
				}
				ct := mimesniffer.Sniff(file)
				return c.Blob(200, ct, file)
			},
		})
		return nil
	})
	pbApp.OnFileDownloadRequest().Add(func(e *pbCore.FileDownloadEvent) error {
		log.Println(e.ServedPath)
		return nil
	})
	return pbApp.Start()
}

// Getting a IPFS node running
func launchIpfs(app *fateCore.App) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Spawn a local peer using a specific repo path
	repoPath := "./data/ipfs/"
	ipfsApi, ipfsNode, err := fateCore.SpawnIpfsNode(ctx, &repoPath)
	if err != nil {
		panic(fmt.Errorf("failed to spawn peer node: %s", err))
	}
	app.IpfsApi = ipfsApi
	app.IpfsNode = ipfsNode

	// Uncomment the following line, only if you need to forcefully add
	// files to the IPFS for testing or any other purpose
	// fateCore.AddFilesToIpfs(ctx, app, "./data/pb/storage/")
	return nil
}
