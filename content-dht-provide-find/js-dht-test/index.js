'use strict'

const libp2p = require('libp2p')
const TCP = require('libp2p-tcp')
const Mplex = require('libp2p-mplex')
const SECIO = require('libp2p-secio')
const PeerInfo = require('peer-info')
const CID = require('cids')
const KadDHT = require('libp2p-kad-dht')
const defaultsDeep = require('@nodeutils/defaults-deep')
const waterfall = require('async/waterfall')
const parallel = require('async/parallel')

class MyBundle extends libp2p {
  constructor (_options) {
    const defaults = {
      modules: {
        transport: [ TCP ],
        streamMuxer: [ Mplex ],
        connEncryption: [ SECIO ],
        // we add the DHT module that will enable Peer and Content Routing
        dht: KadDHT
      },
      config: {
        dht: {
          kBucketSize: 20
        },
        EXPERIMENTAL: {
          dht: true
        }
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
    (cb) => setTimeout(cb, 300)
  ], (err) => {
    if (err) { throw err }

    const provideCid = new CID('QmTp9VkYvnHyrqKQuFPiuZkiX9gPcqj6x5LJ1rmWuSySnL')
    const findCid = new CID('zb2rhXqLbdjpXnJG99QsjM6Nc6xaDKgEr2FfugDJynE7H2NR6')

    node1.contentRouting.provide(provideCid, (err) => {
      if (err) { throw err }

      console.log('Local node %s is providing %s', node1.peerInfo.id.toB58String(), provideCid.toBaseEncodedString())

      setTimeout(() => {
        node1.contentRouting.findProviders(findCid, 10000, (err, providers) => {
          if (err) { throw err }

          if (providers.length !== 0) {
            const provider = providers[0]
            // console.log(provider)
            console.log('Remote node %s is providing %s', provider.id.toB58String(), findCid.toBaseEncodedString())
            setTimeout(() => {
              process.exit(0)
            }, 5000)
          } else {
            console.log('No remote providers found!')
          }
        })
      }, 5000)
    })
  })
})
