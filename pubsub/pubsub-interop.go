package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	//	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	peerstore "github.com/libp2p/go-libp2p-peerstore"

	"github.com/libp2p/go-libp2p-crypto"

	"github.com/libp2p/go-floodsub"
	ma "github.com/multiformats/go-multiaddr"
)

//var ho h.Host
//var dhtPtr *dht.IpfsDHT

func parseArgs() (bool, string) {
	usage := fmt.Sprintf("Usage: %s [-b] [PRIVATE_KEY]\n\n-b is bootstrap mode (creates DHT)\nPRIVATE_KEY is the path to a private key like '../util/private_key.bin'\n", os.Args[0])
	var bBootstrap bool = false
	var privKeyFilePath string
	var args []string = os.Args[1:]
	if (len(args) == 0) || (len(args) > 2) {
		fmt.Printf("Error: wrong number of arguments\n\n%s", usage)
		os.Exit(1)
	}
	if args[0] == "-b" {
		bBootstrap = true
		args = args[1:]
	}
	privKeyFilePath = args[0]
	return bBootstrap, privKeyFilePath
}

func main() {
	ctx := context.Background()

	bBootstrap, privKeyFilePath := parseArgs()
	fmt.Printf("Starting up in ")
	if bBootstrap {
		fmt.Printf("bootstrapper mode")
	} else {
		fmt.Printf("peer mode")
	}
	fmt.Printf(" with private key '%s'", privKeyFilePath)

	//
	// Read the private key
	//
	var privBytes []byte
	privBytes, err := ioutil.ReadFile(privKeyFilePath)
	if err != nil {
		fmt.Println("ioutil.ReadFile:  failed:  %v", err)
		panic(err)
	}

	var priv crypto.PrivKey
	priv, err = crypto.UnmarshalPrivateKey(privBytes)
	if err != nil {
		fmt.Println("crypto.UnmarshalPrivateKey:  failed:  %v", err)
		panic(err)
	}

	//
	// Construct our libp2p host
	//
	host, err := libp2p.New(ctx,
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9876"),
		libp2p.Identity(priv),
	)
	if err != nil {
		fmt.Println("libp2p.New:  failed:  %v", err)
		panic(err)
	}

	//	ho = host
	//fmt.Printf("To connect, run:\n")
	//fmt.Printf("node js-dht-test/index.js %s/ipfs/%s\n", host.Addrs()[0], host.ID().Pretty())

	//
	// Construct a floodsub instance for this host
	//
	fsub, err := floodsub.NewFloodSub(ctx, host)
	if err != nil {
		fmt.Println("Error (floodsub.NewFloodSub): %v", err)
		panic(err)
	}

	//
	// If we are the bootstrap node, don't try to connec to any peers.
	// Else:  try to connect to the bootstrap node.
	//
	const bootstrapAddrIP4Str string = "127.0.0.1"
	if !bBootstrap {
		var bootstrapMultiAddr ma.Multiaddr
		var pinfo *peerstore.PeerInfo
		bootstrapMultiAddrStr := fmt.Sprintf("/ip4/%s/tcp/9876/ipfs/QmehVYruznbyDZuHBV4vEHESpDevMoAovET6aJ9oRuEzWa", bootstrapAddrIP4Str)
		fmt.Println("bootstrapping to '%s'...", bootstrapMultiAddrStr)
		bootstrapMultiAddr, err := ma.NewMultiaddr(bootstrapMultiAddrStr)
		if err != nil {
			fmt.Println("Error (ma.NewMultiaddr): %v", err)
			panic(err)
		}

		pinfo, err = peerstore.InfoFromP2pAddr(bootstrapMultiAddr)
		if err != nil {
			fmt.Println("Error (ma.NewMultiaddr): %v", err)
			panic(err)
		}

		if err := host.Connect(ctx, *pinfo); err != nil {
			fmt.Println("bootstrapping to peer failed: ", err)
		}
	}

	_ = fsub

	// 	//
	// 	// Construct a DHT for peer discovery if we are the bootstrap node.
	// 	// Else:  construct a DHT client for peer discovery and connect to bootstrap node.
	// 	//
	// 	d, err := dht.New(ctx, host, dhtopts.Client(false))
	// 	if err != nil {
	// 		panic(err)
	//	}

	//	dhtPtr = d
}
