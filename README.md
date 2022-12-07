# Faterium Server

Faterium - a place where fates are forged.

Faterium Server built with Golang, embedded IPFS node, and PocketBase. Server auto-imports collections - so developer shouldn't load or create them manually.

## Motivation

We strive for the best UX for the Faterium platform, so from the very beginning, we design the platform not only for Web3 but also for Web2 users. Initially, we planned to store a user profile and a project/community profile on-chain, something like the [Subsocial](https://github.com/dappforce/subsocial-parachain) or [Preimage](https://github.com/paritytech/substrate/tree/master/frame/preimage) approach in Polkadot. But after research, comparisons, and discussions, we realized that the [OpenSea](https://opensea.io/), [Rarible](https://rarible.com/), [Polkassembly](https://polkassembly.io/) approach is more preferable for us, [to minimize storage on the chain](https://wiki.polkadot.network/docs/learn-treasury#announcing-the-proposal).
Therefore, we decided to store data of projects/communities and users' profiles off-chain. Also, we will use the standard `Assets Pallet` to create assets on our platform. But still, we will develop a pallet for `Crowdfunding Polls`, since we have not found anything similar to it.

## 🧞 Commands

Run the following command to launch the server:

```sh
go run ./cmd/main.go serve
```
