# Faterium Server

Faterium - a place where fates are forged.

Faterium Server built with Golang, embedded IPFS node, and PocketBase. Server auto-imports collections - so developer shouldn't load or create them manually.

## Motivation

We strive for the best UX for the Faterium platform, so from the very beginning, we design the platform not only for Web3 but also for Web2 users. Initially, we planned to store a user profile and a project/community profile on-chain, something like the [Subsocial](https://github.com/dappforce/subsocial-parachain) or [Preimage](https://github.com/paritytech/substrate/tree/master/frame/preimage) approach in Polkadot. But after research, comparisons, and discussions, we realized that the [OpenSea](https://opensea.io/), [Rarible](https://rarible.com/), [Polkassembly](https://polkassembly.io/) approach is more preferable for us, [to minimize storage on the chain](https://wiki.polkadot.network/docs/learn-treasury#announcing-the-proposal).
Therefore, we decided to store data of projects/communities and users' profiles off-chain. Also, we will use the standard `Assets Pallet` to create assets on our platform. But still, we will develop a pallet for `Crowdfunding Polls`, since we have not found anything similar to it.

## Commands

Run the following command to launch the server:

```sh
go run ./cmd/main.go serve
```

## Docker and local network

There're currently two infrastructure setups for `faterium-dapp` server: `local` and `testnet`. We recommend to start with `local` to see how to launch server, node, and polkadot-apps locally.

To launch local docker containers - just run the following command:

```sh
docker-compose up
```

It will pull latest `faterium-node` image from github docker package registry, build local golang PocketBase IPFS server inside the image, and launch services on the following ports: `faterium-node` on `localhost:9944`, `faterium-server` on `localhost:8090` (with `/api` and `/_` endpoints), `polkadot-apps` on `localhost:8000`.

To rebuild the local image from source code you must use `docker-compose build` or `docker-compose up --build`.

To clean up docker containers:

```sh
docker-compose down
```

If you want to clean up the blocks and server data - delete `./data` folder.

## Dedicated testnet server

If you want to run all these services on the dedicated server and behind `https://` and `wss://` - feel free to use our setup from `infra/testnet` folder. It's almost the same as the local setup but it uses [traefik reverse-proxy](https://traefik.io/) to setup services on our `dapp-*.faterium.com` subdomains (you can change them in `docker-compose.yml` file) and `SSL/TLS` encryption. Feel free to read its [docs](https://doc.traefik.io/traefik/) and our `docker-compose.yml` in the `infra/testnet` folder.
