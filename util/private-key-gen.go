package main

import (
	"fmt"
	"io/ioutil"
	"os"

	_ "github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-crypto"
	_ "github.com/libp2p/go-libp2p-host"
	_ "github.com/libp2p/go-libp2p-kad-dht"
	_ "github.com/libp2p/go-libp2p-kad-dht/opts"
	_ "github.com/libp2p/go-libp2p-net"
)

func main() {
	var priv crypto.PrivKey
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 4096)
	if err != nil {
		fmt.Println("crypto.GenerateKeyPair:  failed:  %v", err)
		panic(err)
	}

	var privBytes []byte
	privBytes, err = crypto.MarshalPrivateKey(priv)
	if err != nil {
		fmt.Println("crypto.MarshalPrivateKey:  failed:  %v", err)
		panic(err)
	}

	// Print the marshalled bytes
	//n := len(privBytes)
	//s := string(privBytes[:n])
	//fmt.Printf("*** <%s> (n=%v)\n", s, n)

	ioutil.WriteFile("private_key.bin", privBytes, os.ModePerm)
	if err != nil {
		fmt.Println("ioutil.WriteFile:  failed:  %v", err)
		panic(err)
	}
}
