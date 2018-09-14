package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	_ "time"

	//	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peerstore "github.com/libp2p/go-libp2p-peerstore"

	"github.com/libp2p/go-libp2p-crypto"

	"github.com/libp2p/go-floodsub"
	ma "github.com/multiformats/go-multiaddr"
)

var ho host.Host

//var dhtPtr *dht.IpfsDHT

//var TopicName string = "libp2p-demo-chat"
var TopicName string = "RDEpsjSPrAZF9JCK5REt3tao"

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

func handleConn(conn inet.Conn) {
	ctx := context.Background()
	h := ho
	fmt.Printf("<NOTICE> Got connection from %v\n", conn.RemoteMultiaddr().String())
	_ = h
	_ = ctx
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
	fmt.Printf(" with private key '%s'\n", privKeyFilePath)

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
	var host host.Host
	if bBootstrap {
		host, err = libp2p.New(ctx,
			libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9876"),
			libp2p.Identity(priv),
		)
	} else {
		host, err = libp2p.New(ctx,
			libp2p.Identity(priv),
		)
	}
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
		fmt.Printf("bootstrapping to '%s'...\n", bootstrapMultiAddrStr)
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

	//
	// Subscribe to the topic and wait for messages published on that topic
	//
	sub, err := fsub.Subscribe(TopicName)
	if err != nil {
		fmt.Println("Error (fsub.Subscribe): %v", err)
		panic(err)
	}

	// Go and listen for messages from them, and print them to the screen
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				fmt.Println("Error (sub.Next): %v", err)
				panic(err)
			}

			fmt.Printf("%s: %s\n", msg.GetFrom(), string(msg.GetData()))
		}
	}()

	// SetConnHandler() should not normally be called.  Instead,
	// use Notify() and pass it a functioon.
	// The problem with SetConnHandler() is that it takes control
	// of the connection.
	//host.Network().SetConnHandler(handleConn)
	host.Network().Notify(&inet.NotifyBundle{
		ConnectedF: func(n inet.Network, c inet.Conn) {
			fmt.Println("Got a connection:", c.RemotePeer())
		},
	})
	if bBootstrap {
		fmt.Println("Bootstrapper running.  Ctrl+C to exit.")
		for true {
		}
	} else {
		// Now, wait for input from the user, and send that out!
		fmt.Println("Type something and hit enter to send to other subscribers:")
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			if err := fsub.Publish(TopicName, scan.Bytes()); err != nil {
				panic(err)
			}
		}
	}
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
