package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	h "github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/opts"
	"github.com/libp2p/go-libp2p-net"

	"github.com/libp2p/go-libp2p-crypto"
)

var ho h.Host
var dhtPtr *dht.IpfsDHT

func handleConn(conn net.Conn) {
	ctx := context.Background()

	d := *dhtPtr

	provideCid, err := cid.Decode("zb2rhXqLbdjpXnJG99QsjM6Nc6xaDKgEr2FfugDJynE7H2NR6")
	if err != nil {
		panic(err)
	}
	findCid, err := cid.Decode("QmTp9VkYvnHyrqKQuFPiuZkiX9gPcqj6x5LJ1rmWuSySnL")
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	// First, announce ourselves as participating in this topic
	fmt.Println("announcing ourselves...")
	tctx, _ := context.WithTimeout(ctx, time.Second*10)
	if err := d.Provide(tctx, provideCid, true); err != nil {
		panic(err)
	}

	fmt.Printf("Local node %s is providing %s\n", ho.ID().Pretty(), provideCid)

	// Now, look for others who have announced
	fmt.Println("searching for other peers...")
	tctx, _ = context.WithTimeout(ctx, time.Second*10)
	providers, err := d.FindProviders(tctx, findCid)
	if err != nil {
		panic(err)
	}

	if len(providers) != 0 {
		provider := providers[0]
		fmt.Printf("Remote node %s is providing %s\n", provider.ID.Pretty(), findCid)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	} else {
		fmt.Printf("no remote providers!\n")
	}
}

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
	if !bBootstrap {
		fmt.Printf("Error: bootstrap mode required in this example")
		os.Exit(1)
	}

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
	// Set up a libp2p host.
	//
	host, err := libp2p.New(ctx,
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9876"),
		libp2p.Identity(priv),
	)
	if err != nil {
		fmt.Println("libp2p.New:  failed:  %v", err)
		panic(err)
	}

	host.Network().SetConnHandler(handleConn)

	ho = host

	fmt.Printf("To connect, run:\n")
	fmt.Printf("node js-dht-test/index.js %s/ipfs/%s\n", host.Addrs()[0], host.ID().Pretty())

	//
	// Construct a DHT for discovery.
	//
	d, err := dht.New(ctx, host, dhtopts.Client(false))
	if err != nil {
		panic(err)
	}

	dhtPtr = d

	select {}
}
