package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbCore "github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

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
	dataDirPath := "./data/pb/"
	pbApp := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDebug:   false,
		DefaultDataDir: dataDirPath,
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
	pbApp.OnRecordBeforeCreateRequest().Add(func(e *pbCore.RecordCreateEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName != "images" {
			return nil
		}
		form, err := e.HttpContext.MultipartForm()
		if err != nil {
			return err
		}
		mapFile := form.File["file"][0]
		if mapFile == nil {
			return fmt.Errorf("invalid request file")
		}
		file, err := mapFile.Open()
		if err != nil {
			return err
		}
		res, err := fateCore.AddBytesToIpfs(e.HttpContext.Request().Context(), app, file)
		if err != nil {
			return err
		}
		// Update Record cid field with the actual IPFS CID of the file
		e.Record.Set("cid", res.Cid().String())
		return nil
	})
	pbApp.RootCmd.AddCommand(&cobra.Command{
		Use: "syncIpfs",
		Run: func(command *cobra.Command, args []string) {
			// The following line used only if you need to forcefully add
			// files to the IPFS for testing or any other purpose
			fateCore.AddFilesToIpfs(command.Context(), app, dataDirPath+"storage/")
		},
	})
	return pbApp.Start()
}

// Getting a IPFS node running
func launchIpfs(app *fateCore.App) error {
	// Spawn a local peer using a specific repo path
	repoPath := "./data/ipfs/"
	ipfsApi, ipfsNode, err := fateCore.SpawnIpfsNode(context.Background(), &repoPath)
	if err != nil {
		return fmt.Errorf("failed to spawn peer node: %s", err)
	}
	app.IpfsApi = ipfsApi
	app.IpfsNode = ipfsNode
	return nil
}
