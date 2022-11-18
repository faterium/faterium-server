package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbCore "github.com/pocketbase/pocketbase/core"

	files "github.com/ipfs/go-ipfs-files"
	ipfsPath "github.com/ipfs/interface-go-ipfs-core/path"

	fateCore "github.com/faterium/faterium-server/core"
)

func main() {
	log.Print("let's get runnin' this server!")

	// Server App struct to controll services
	app := fateCore.App{}
	// Add IPFS node service
	app.AddService("ipfs", launchIPFS)
	// Add PocketBase service
	app.AddService("pocketbase", launchPocketBase)
	// Launch all services
	app.LaunchServices()
}

func launchPocketBase(app *fateCore.App) error {
	pbApp := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDebug: false,
	})
	pbApp.OnBeforeServe().Add(func(e *pbCore.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/ipfs/:cid",
			Handler: func(c echo.Context) error {
				cid := c.PathParam("cid")
				file, err := GetFileFromIpfs(c.Request().Context(), app, cid)
				if err != nil {
					return err
				}
				ctVideo := "video/mp4, video/x-ms-wmv, video/quicktime, video/3gpp"
				ctImage := "image/jpg, image/jpeg, image/png, image/svg+xml, image/gif"
				contentType := ctVideo + ", " + ctImage
				return c.Blob(200, contentType, file)
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
func launchIPFS(app *fateCore.App) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Spawn a local peer using a temporary path
	ipfsApi, ipfsNode, err := fateCore.SpawnIpfsNode(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to spawn peer node: %s", err))
	}
	app.IpfsApi = ipfsApi
	app.IpfsNode = ipfsNode

	// TODO: Remove this line, only needed for testing
	SetupIpfs(ctx, app)
	return nil
}

func SetupIpfs(ctx context.Context, app *fateCore.App) error {
	err := filepath.Walk("./pb_data/storage/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Adding a file to IPFS
			cid, err := AddFileToIpfs(ctx, app, path)
			if err != nil {
				return fmt.Errorf("could not get File: %s", err)
			}
			fmt.Println(path, info.Size(), cid)
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return nil
}

// Add a file or a directory to IPFS
func AddFileToIpfs(ctx context.Context, app *fateCore.App, inputPath string) (ipfsPath.Resolved, error) {
	inputNode, err := fateCore.GetUnixfsNode(inputPath)
	if err != nil {
		return nil, fmt.Errorf("could not get File: %s", err)
	}
	cid, err := app.IpfsApi.Unixfs().Add(ctx, inputNode)
	if err != nil {
		return nil, fmt.Errorf("could not add Directory: %s", err)
	}
	return cid, nil
}

// Get the file that was added once to IPFS
func GetFileFromIpfs(ctx context.Context, app *fateCore.App, cid string) ([]byte, error) {
	fetchCid := "/ipfs/" + cid
	ipfsCID := ipfsPath.New(fetchCid)

	node, err := app.IpfsApi.Unixfs().Get(ctx, ipfsCID)
	if err != nil {
		return nil, fmt.Errorf("could not get file with CID: %s", err)
	}

	buf := new(bytes.Buffer)
	switch nd := node.(type) {
	case files.File:
		_, err := io.Copy(buf, nd)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("requested with invalid cid")
	}
	return buf.Bytes(), nil
}
