// Copyright 2018 Parity Technologies (UK) Ltd.
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
// OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.
use futures::prelude::*;
use libp2p::{
    NetworkBehaviour, Transport,
    core::upgrade::{self, OutboundUpgradeExt},
    secio,
    mplex,
    multiaddr,
    tokio_codec::{FramedRead, LinesCodec}
};
use std::env;

fn main() {

    // Create a random PeerId
    let local_key = secio::SecioKeyPair::ed25519_generated().unwrap();
    let local_pub_key = local_key.to_public_key();

    // Set up a an encrypted DNS-enabled TCP Transport over the Mplex protocol
    let transport = libp2p::CommonTransport::new()
        .with_upgrade(secio::SecioConfig::new(local_key))
        .and_then(move |out, _| {
            let peer_id = out.remote_key.into_peer_id();
            let upgrade = mplex::MplexConfig::new().map_outbound(move |muxer| (peer_id, muxer) );
            upgrade::apply_outbound(out.stream, upgrade).map_err(|e| e.into_io_error())
        });

    // Create a Floodsub topic
    let floodsub_topic = libp2p::floodsub::TopicBuilder::new("libp2p-demo-chat").build();

    // We create a custom network behaviour that combines floodsub and mDNS.
    // In the future, we want to improve libp2p to make this easier to do.
    #[derive(NetworkBehaviour)]
    struct MyBehaviour<TSubstream: libp2p::tokio_io::AsyncRead + libp2p::tokio_io::AsyncWrite> {
        #[behaviour(handler = "on_floodsub")]
        floodsub: libp2p::floodsub::Floodsub<TSubstream>,
        mdns: libp2p::mdns::Mdns<TSubstream>,
    }

    impl<TSubstream: libp2p::tokio_io::AsyncRead + libp2p::tokio_io::AsyncWrite> MyBehaviour<TSubstream> {
        // Called when `floodsub` produces an event.
        fn on_floodsub<TTopology>(&mut self, message: <libp2p::floodsub::Floodsub<TSubstream> as libp2p::core::swarm::NetworkBehaviour<TTopology>>::OutEvent)
        where TSubstream: libp2p::tokio_io::AsyncRead + libp2p::tokio_io::AsyncWrite
        {
            println!("{:?}: {}", &message.source,String::from_utf8_lossy(&message.data));
        }
    }

    // Create a Swarm to manage peers and events
    let mut swarm = {
        let mut behaviour = MyBehaviour {
            floodsub: libp2p::floodsub::Floodsub::new(local_pub_key.clone().into_peer_id()),
            mdns: libp2p::mdns::Mdns::new().expect("Failed to create mDNS service"),
        };

        behaviour.floodsub.subscribe(floodsub_topic.clone());
        libp2p::Swarm::new(transport, behaviour, libp2p::core::topology::MemoryTopology::empty(), local_pub_key)
    };

    // We dial a bootstrap node for the nodes outside of the reach of mDNS.
    libp2p::Swarm::dial_addr(&mut swarm, "/ip4/127.0.0.1/tcp/5555".parse().unwrap()).unwrap();

    // Listen on all interfaces.
    let port = if let Some(port) = env::args().nth(1) {
        port.parse().expect("Failed to parse port number")
    } else {
        0u16
    };

    let address = libp2p::Swarm::listen_on(&mut swarm, multiaddr![Ip4([0, 0, 0, 0]), Tcp(port)]).unwrap();
    println!("Now listening on {:?}", address);

    // Read full lines from stdin
    println!("Type your message to send to remote hosts:");
    let stdin = tokio_stdin_stdout::stdin(0);
    let mut framed_stdin = FramedRead::new(stdin, LinesCodec::new()).fuse();

    // Kick it off
    tokio::run(futures::future::poll_fn(move || -> Result<_, ()> {
        loop {
            match framed_stdin.poll().expect("Error while polling stdin") {
                Async::Ready(Some(line)) => {
                    let to_send = format!("{}", line);
                    println!("sending: {}", line);
                    swarm.floodsub.publish(&floodsub_topic, to_send.as_bytes())
                },
                Async::Ready(None) => break, // Stdin closed
                Async::NotReady => break,
            };
        }

        loop {
            match swarm.poll().expect("Error while polling swarm") {
                Async::Ready(Some(_)) => {
                },
                Async::Ready(None) | Async::NotReady => break,
            }
        }

        Ok(Async::NotReady)
    }));
}
