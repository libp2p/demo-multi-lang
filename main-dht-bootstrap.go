package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"
	"strings"

	"github.com/multiformats/go-multihash"
	"github.com/libp2p/go-floodsub"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/opts"
	_ "github.com/libp2p/go-libp2p-peerstore"
	"github.com/ipfs/go-cid"
//	"github.com/ipfs/go-datastore"
	_ "github.com/ipfs/go-ipfs-addr"
)

const bootstrapHostname string = "libp2p-bootstrap.goelzer.io"

func usage() {
	fmt.Printf("Usage: %s [OPTIONS]\n\n", os.Args[0]);
	fmt.Printf("  --bootstrapper   Start a DHT; initial peer only.\n");
	fmt.Printf("  --peer=IP        Connect to IP to join the libp2p swarm\n");
	fmt.Printf("  --port=N         Listen and connect on N\n");
	fmt.Printf("  --help           Display this message\n");
	fmt.Printf("Note that --bootstrapper and --peer are mutually exclusive.\n");
	fmt.Printf("\nThis program demonstrates a libp2p swarm. The first node\n");
	fmt.Printf("is started with `--bootstrapper`.  Add a second node with\n");
	fmt.Printf("`--peer=IP_OF_FIRST_NODE`.\n\n");
}

func printMutuallyExclusiveErrorAndDie() {
	fmt.Printf("Error: --bootstrapper and --peer are mutually exclusive\n\n");
	os.Exit(1);
}

func parseArgs(isBootstrap *bool, peerAddrStr *string) {
	*isBootstrap = false;
	*peerAddrStr = "";

	for _,arg := range os.Args[1:] {
		if (arg == "--help") {
			// --help = print usage and die
			usage();
			os.Exit(1);
		} else if (arg == "--bootstrapper") {
			// Bootstrap mode '--bootstrapper' => we create a DHT
			if (*peerAddrStr != "") {
				printMutuallyExclusiveErrorAndDie();
			}
			*isBootstrap = true;
		} else if (strings.HasPrefix(arg,"--peer=")) {
			// Peer mode:  won't create a DHT but instead connect to peer IP
			if (*isBootstrap == true) {
				printMutuallyExclusiveErrorAndDie();
			}
			*peerAddrStr = arg[7:];
		} else {
			fmt.Printf("Invalid argument: '%s'\n\n",arg);
			os.Exit(1);
		}
	}
}

func main() {
	var isBootstrap bool;
	var peerAddrStr string;
	parseArgs(&isBootstrap, &peerAddrStr);

	if (isBootstrap) {
		fmt.Println("Bootstrap Mode");
	} else {
		fmt.Printf("Peer Mode (peer address = '%s')\n",peerAddrStr);
	}

	ctx := context.Background()

	//
	// Set up a libp2p host.
	//
	host, err := libp2p.New(ctx, libp2p.Defaults)
	if err != nil {
		fmt.Println("libp2p.New:  failed:  %v",err)
		panic(err)
	}
	// TODO:  get my own public IP instead of assuming 159.89.221.55 (IP of libp2p-bootstrap.goelzer.io)
	fmt.Println("My adress:  /ip4/159.89.221.55/tcp/5555/ipfs/%s", host.ID().Pretty() )

	// TODO:  rename to PubSubTopicName
	TopicName := "libp2p-go-js-rust-chat"
	_ = TopicName

	//
	// Construct ourselves a pubsub instance using that libp2p host.
	//
	fsub, err := floodsub.NewFloodSub(ctx, host)
	if err != nil {
		panic(err)
	}

	_ = fsub

	//
	// Construct a DHT for discovery.
	//
	dht, err := dht.New(ctx, host, dhtopts.Client(false) )
	if err != nil {
		panic(err)
	}

	_ = dht

// These are the IPFS bootstrap nodes:
//
//	bootstrapPeers := []string{
//		"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
//		"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
//		"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
//		"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
//		"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
//	}


////	bootstrapPeers := []string{
////		fmt.Sprintf("/ip4/159.89.221.55/tcp/5555/ipfs/%s", host.ID().Pretty()),
////	}

//
//	fmt.Println("bootstrapping...")
//	for _, addr := range bootstrapPeers {
//		iaddr, _ := ipfsaddr.ParseString(addr)
//
//		pinfo, _ := peerstore.InfoFromP2pAddr(iaddr.Multiaddr())
//
//		if err := host.Connect(ctx, *pinfo); err != nil {
//			fmt.Println("bootstrapping to peer failed: ", err)
//		}
//	}

	// Using the sha256 of our "topic" as our rendezvous value
	c, _ := cid.NewPrefixV1(cid.Raw, multihash.SHA2_256).Sum([]byte(TopicName))

	// First, announce ourselves as participating in this topic
	fmt.Println("announcing ourselves...")
	tctx, _ := context.WithTimeout(ctx, time.Second*10)
	if err := dht.Provide(tctx, c, true); err != nil {
		panic(err)
	}

//	// Now, look for others who have announced
//	fmt.Println("searching for other peers...")
//	tctx, _ = context.WithTimeout(ctx, time.Second*10)
//	peers, err := dht.FindProviders(tctx, c)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Printf("Found %d peers!\n", len(peers))
//
//	// Now connect to them!
//	for _, p := range peers {
//		if p.ID == host.ID() {
//			// No sense connecting to ourselves
//			continue
//		}
//
//		tctx, _ := context.WithTimeout(ctx, time.Second*5)
//		if err := host.Connect(tctx, p); err != nil {
//			fmt.Println("failed to connect to peer: ", err)
//		}
//	}
//
//	fmt.Println("bootstrapping and discovery complete!")
//
	sub, err := fsub.Subscribe(TopicName)
	if err != nil {
		panic(err)
	}

	// Go and listen for messages from them, and print them to the screen
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s: %s\n", msg.GetFrom(), string(msg.GetData()))
		}
	}()

	// Now, wait for input from the user, and send that out!
	fmt.Println("Type something and hit enter to send:")
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		if err := fsub.Publish(TopicName, scan.Bytes()); err != nil {
			panic(err)
		}
	}











}

