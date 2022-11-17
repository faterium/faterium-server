package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pocketbase/pocketbase"
	pbCore "github.com/pocketbase/pocketbase/core"

	files "github.com/ipfs/go-ipfs-files"
)

func main() {
	log.Print("let's get runnin'")

	var wg sync.WaitGroup

	// Launch PocketBase server
	wg.Add(1)
	go func() {
		defer wg.Done()
		launchPocketBase()
	}()

	// Launch IPFS node
	wg.Add(1)
	go func() {
		defer wg.Done()
		launchIPFS()
	}()

	wg.Wait()
}

func launchPocketBase() {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDebug: false,
	})

	app.OnFileDownloadRequest().Add(func(e *pbCore.FileDownloadEvent) error {
		log.Println(e.ServedPath)
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func launchIPFS() {
	flag.Parse()

	/// Getting a IPFS node running
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Spawn a local peer using a temporary path, for testing purposes
	ipfsA, _, err := spawnEphemeral(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to spawn peer node: %s", err))
	}

	// peerCidFile, err := ipfsA.Unixfs().Add(ctx,
	// 	files.NewBytesFile([]byte("hello from ipfs 101 in Kubo")))
	// if err != nil {
	// 	panic(fmt.Errorf("could not add File: %s", err))
	// }

	fmt.Println("IPFS node is running")

	/// Adding a file and a directory to IPFS

	inputBasePath := "./pb_data/storage"
	inputPathDirectory := inputBasePath

	someDirectory, err := getUnixfsNode(inputPathDirectory)
	if err != nil {
		panic(fmt.Errorf("could not get File: %s", err))
	}

	cidDirectory, err := ipfsA.Unixfs().Add(ctx, someDirectory)
	if err != nil {
		panic(fmt.Errorf("could not add Directory: %s", err))
	}

	fmt.Printf("Added directory to IPFS with CID %s\n", cidDirectory.String())

	/// Getting the file and directory you added back

	outputBasePath := "./ipfs/"
	if err := os.RemoveAll(outputBasePath); err != nil {
		panic(fmt.Errorf("failed to remove ipfs dir: %s", err))
	}
	if err := os.MkdirAll(outputBasePath, 0700); err != nil {
		panic(fmt.Errorf("failed to create ipfs dir: %s", err))
	}
	outputPathDirectory := outputBasePath + strings.Split(cidDirectory.String(), "/")[2]

	rootNodeDirectory, err := ipfsA.Unixfs().Get(ctx, cidDirectory)
	if err != nil {
		panic(fmt.Errorf("could not get file with CID: %s", err))
	}

	err = files.WriteTo(rootNodeDirectory, outputPathDirectory)
	if err != nil {
		panic(fmt.Errorf("could not write out the fetched CID: %s", err))
	}

	fmt.Printf("Got directory back from IPFS (IPFS path: %s) and wrote it to %s\n", cidDirectory.String(), outputPathDirectory)

	fmt.Println("\nAll done! You just finalized your first tutorial on how to use Kubo as a library")
}
