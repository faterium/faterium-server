package core

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	files "github.com/ipfs/go-ipfs-files"
	icore "github.com/ipfs/interface-go-ipfs-core"
	ipfsPath "github.com/ipfs/interface-go-ipfs-core/path"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/plugin/loader" // This package is needed so that all the preloaded plugins are loaded automatically
	"github.com/ipfs/kubo/repo/fsrepo"
	"github.com/libp2p/go-libp2p/core/peer"
)

var flagExp = flag.Bool("experimental", false, "enable experimental features")

func SetupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func CreateRepo(path *string) (string, error) {
	repoPath := *path
	if path == nil {
		var err error
		repoPath, err = os.MkdirTemp("", "ipfs-shell")
		if err != nil {
			return "", fmt.Errorf("failed to get temp dir: %s", err)
		}
	}

	// Create a config with default options and a 2048 bit key
	cfg, err := config.Init(io.Discard, 2048)
	if err != nil {
		return "", err
	}

	// When creating the repository, you can define custom settings on the repository, such as enabling experimental
	// features (See experimental-features.md) or customizing the gateway endpoint.
	// To do such things, you should modify the variable `cfg`. For example:
	if *flagExp {
		// https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md#ipfs-filestore
		cfg.Experimental.FilestoreEnabled = true
		// https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md#ipfs-urlstore
		cfg.Experimental.UrlstoreEnabled = true
		// https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md#ipfs-p2p
		cfg.Experimental.Libp2pStreamMounting = true
		// https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md#p2p-http-proxy
		cfg.Experimental.P2pHttpProxy = true
		// See also: https://github.com/ipfs/kubo/blob/master/docs/config.md
		// And: https://github.com/ipfs/kubo/blob/master/docs/experimental-features.md
	}

	// Create the repo with the config
	err = fsrepo.Init(repoPath, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to init ephemeral node: %s", err)
	}

	return repoPath, nil
}

// Creates an IPFS node and returns its coreAPI
func CreateNode(ctx context.Context, repoPath string) (*core.IpfsNode, error) {
	// Open the repo
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, err
	}

	// Construct the node
	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		// Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo: repo,
	}

	return core.NewNode(ctx, nodeOptions)
}

var loadPluginsOnce sync.Once

// Parses flags and spawns IPFS node
func SpawnIpfsNode(ctx context.Context, pathToRepo *string) (icore.CoreAPI, *core.IpfsNode, error) {
	flag.Parse()

	var onceErr error
	loadPluginsOnce.Do(func() {
		onceErr = SetupPlugins("")
	})
	if onceErr != nil {
		return nil, nil, onceErr
	}

	// Create a Temporary Repo
	repoPath, err := CreateRepo(pathToRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp repo: %s", err)
	}

	node, err := CreateNode(ctx, repoPath)
	if err != nil {
		return nil, nil, err
	}

	api, err := coreapi.NewCoreAPI(node)

	return api, node, err
}

func ConnectToPeers(ctx context.Context, ipfs icore.CoreAPI, peers []string) error {
	var wg sync.WaitGroup
	peerInfos := make(map[peer.ID]*peer.AddrInfo, len(peers))
	for _, addrStr := range peers {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		pi, ok := peerInfos[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			peerInfos[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	wg.Add(len(peerInfos))
	for _, peerInfo := range peerInfos {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := ipfs.Swarm().Connect(ctx, *peerInfo)
			if err != nil {
				log.Printf("failed to connect to %s: %s", peerInfo.ID, err)
			}
		}(peerInfo)
	}
	wg.Wait()
	return nil
}

func GetUnixfsNode(path string) (files.Node, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	f, err := files.NewSerialFile(path, false, st)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Get the file that was added once to IPFS
func GetFileFromIpfs(ctx context.Context, app *App, cid string) ([]byte, error) {
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

// Recursively add all files in the path to IPFS
func AddFilesToIpfs(ctx context.Context, app *App, path string) error {
	fmt.Printf("adding all files by path \"%s\" to IPFS node", path)
	err := filepath.Walk(path,
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
func AddFileToIpfs(ctx context.Context, app *App, inputPath string) (ipfsPath.Resolved, error) {
	inputNode, err := GetUnixfsNode(inputPath)
	if err != nil {
		return nil, fmt.Errorf("could not get File: %s", err)
	}
	cid, err := app.IpfsApi.Unixfs().Add(ctx, inputNode)
	if err != nil {
		return nil, fmt.Errorf("could not add Directory: %s", err)
	}
	return cid, nil
}
