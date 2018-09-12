'use strict'

const libp2p = require('libp2p')
const TCP = require('libp2p-tcp')
const Mplex = require('libp2p-mplex')
const SECIO = require('libp2p-secio')
const PeerInfo = require('peer-info')
const FloodSub = require('libp2p-floodsub')
const CID = require('cids')
const KadDHT = require('libp2p-kad-dht')
const defaultsDeep = require('@nodeutils/defaults-deep')
const waterfall = require('async/waterfall')
const parallel = require('async/parallel')

var fsub;

class MyBundle extends libp2p {
  constructor(_options) {
    const defaults = {
      modules: {
        transport: [TCP],
        streamMuxer: [Mplex],
        connEncryption: [SECIO],
        //        // we add the DHT module that will enable Peer and Content Routing
        //        dht: KadDHT
      },
      config: {
        //        dht: {
        //          kBucketSize: 20
        //        },
        //        EXPERIMENTAL: {
        //          dht: true
        //        }
      }
    }

    super(defaultsDeep(_options, defaults))
  }
}

function createNode(callback) {
  let node

  waterfall([
    (cb) => PeerInfo.create(cb),
    (peerInfo, cb) => {
      peerInfo.multiaddrs.add('/ip4/0.0.0.0/tcp/0')
      node = new MyBundle({
        peerInfo
      })
      node.start(cb)
    }
  ], (err) => callback(err, node))
}

var node;
const bootstrapAddr = process.argv[2];
waterfall([
  (cb) => createNode(cb),
  (node_, cb) => {
    node = node_
    console.log("My ID:  " + node.peerInfo.id._idB58String)
    fsub = new FloodSub(node)
    fsub.start(cb)
  },
  (cb) => {
    fsub.on('libp2p-demo-chat', (data) => {
      const peerIdStr = data.from
      const peerIdTruncdStr = peerIdStr.substr(0,2) + "*" + peerIdStr.substr(peerIdStr.length-6,6)
      const messageStr = data.data
      console.log("<peer " + peerIdTruncdStr + ">: " + messageStr)
    })
    fsub.subscribe('libp2p-demo-chat')

    node.dial(bootstrapAddr, cb)
  }
], (err) => {
  if (err) {
    console.log('Error:', err)
    throw err
  }
  console.log("Connected to:", bootstrapAddr)
  setInterval(pubsubloop, 3000);
})


var i = 0

function pubsubloop() {
  i = i + 1
  var s = new Buffer('Hello from JS (' + i + ')')
  fsub.publish('libp2p-demo-chat', s)
}