

util/private-key-gen -> private_key.bin (marshalled private key that is used for stable peer id of bootstrap server)



0.  Help me fix node "Error: Cannot find module 'libp2p'"
Answer:
  cd js-dht-test
  npm install


2.  Let's run it on bootstrap box (go) and locally (js)
	WORKS!

3.  Explain to me how you generated peer CIDs
	One is the hash of a pubsub topic
	The other is just mine

All this is example does is verify that two peers can find each other's content

