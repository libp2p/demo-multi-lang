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

// "RDE...tao" is the hash of "libp2p-chat-demo" (for compat with Rust)
const topicName = "RDEpsjSPrAZF9JCK5REt3tao"

class MyBundle extends libp2p {
  constructor(_options) {
    const defaults = {
      modules: {
        transport: [TCP],
        streamMuxer: [Mplex],
        connEncryption: [SECIO],
      },
    }

    super(defaultsDeep(_options, defaults))
  }
}

function createNode(callback) {
  let node

  waterfall([
    (cb) => PeerInfo.create(cb),
    (peerInfo, cb) => {
      peerInfo.multiaddrs.add('/ip4/0.0.0.0/tcp/0')  //TODO:  0 -> 6001 ?
      node = new MyBundle({
        peerInfo
      })
      node.start(cb)
    }
  ], (err) => callback(err, node))
}

let fsub;
let node;
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
    fsub.on(topicName, (data) => {
      const peerIdStr = data.from
      const peerIdTruncdStr = peerIdStr.substr(0,2) + "*" + peerIdStr.substr(peerIdStr.length-6,6)
      const messageStr = data.data
      console.log("<peer " + peerIdTruncdStr + ">: " + messageStr)
    })
    fsub.subscribe(topicName)

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


let i = 0
function pubsubloop() {
  i = i + 1
  var s = new Buffer('Hello from JS (' + i + ')')
  fsub.publish(topicName, s)
}