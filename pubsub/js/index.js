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
  constructor (_options) {
    const defaults = {
      modules: {
        transport: [ TCP ],
        streamMuxer: [ Mplex ],
        connEncryption: [ SECIO ],
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

function createNode (callback) {
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

parallel([
  (cb) => createNode(cb),
], (err, nodes) => {
  if (err) { throw err }

  const node1 = nodes[0]

  const bootstrapAddr = process.argv[2]
  console.log('Connecting to:', bootstrapAddr)

  parallel([
    (cb) => node1.dial(bootstrapAddr, cb),
    // Set up of the cons might take time
    (cb) => setTimeout(cb, 3000)
  ], (err) => { if (err) { throw err }


    setTimeout( function() {
      console.log("My ID:  " + node1.peerInfo.id._idB58String)

      fsub = new FloodSub(node1)
      console.log("Initialized fsub\n")
      fsub.start((err) => {
        if (err) {
          console.log('Error:', err)
        }
        fsub.on('libp2p-demo-chat', (data) => {
          console.log(">>> Got some data:")
          console.log("  From:     " + data.from);
          console.log("  Data:     " + data.data);
          console.log("  SeqNo:    " + data.seqno);
          console.log("  topicIDs: " + data.topicIDs);
          console.log(">>> END -- data\n")
        })
        fsub.subscribe('libp2p-demo-chat')
      
        fsub.publish('libp2p-demo-chat', new Buffer('Hello from JS (0)'))
      })
    }, 3000)

  })
})


var i = 0
function pubsubloop() {
    i = i + 1
    var s = new Buffer('Hello from JS ('+i+')')
    fsub.publish('libp2p-demo-chat',s)
}

setInterval(pubsubloop,10000);