#![allow(dead_code)]
extern crate mio;

use self::mio::channel::{channel, Sender, Receiver};
use self::mio::{Poll, Ready, Token, PollOpt, Events};
use network::Connection;
use node::{NodeCommand, Event, NodeCMD, EVENT_ON_CONNECTION_OPEN};
use std::collections::BTreeMap;
use network::tcp::{TcpNetwork, TcpNetworkCommand, TcpNetworkCMD};
use std::thread;
use std::process;
use std::sync::Arc;

const RECEIVER_CHANNEL_TOKEN: Token = Token(0);

/// Base structure for handling all Networking actions for this project
pub struct Network {
    // BTreeMap for keeping
    // key -> connection unique prime value
    // value -> Network connection object
    connections: BTreeMap<u64, Connection>,

    // Sender and Receiver for handling commands for Networking
    sender_channel: Sender<NetworkCommand>,
    receiver_channel: Receiver<NetworkCommand>,

    // channel for base Node service
    node_channel: Sender<NodeCommand>,

    // address for binding TCP server
    tcp_address: String,

    // channel for sending commands to TCP networking
    tcp_channel: Vec<Sender<TcpNetworkCommand>>,

    // poll handler for base Networking
    poll: Poll,

    // keeping current node API version for sending during client requests
    current_api_version: u32,

    // keeping current node Value for sending it during client requests
    current_value: u64
}

/// Enumeration for commands available for Networking
pub enum NetworkCMD {
    HandleNewConnection,
    HandleEventData,
    TCPClientConnection
}

/// Base structure for transferring command over loops to Networking
pub struct NetworkCommand {
    pub cmd: NetworkCMD,
    pub connection: Vec<Connection>,
    pub event: Vec<Event>,
    pub client_address: String
}

impl Network {
    pub fn new(tcp_address: &str, node_chan: Sender<NodeCommand>, api_version: u32, current_value: u64) -> Network {
        let (s, r) = channel::<NetworkCommand>();
        Network {
            connections: BTreeMap::new(),
            sender_channel: s,
            receiver_channel: r,
            tcp_address: String::from(tcp_address),
            poll: Poll::new().expect("Unable to make POLL service for base networking !"),
            node_channel: node_chan,
            current_api_version: api_version,
            current_value: current_value,
            tcp_channel: vec![]
        }
    }

    #[inline(always)]
    pub fn channel(&self) -> Sender<NetworkCommand> {
        self.sender_channel.clone()
    }

    /// Starting base Networking service
    /// this will start also TCP networking and his own POLL service
    pub fn start(&mut self, concurrency: usize) {
        // starting TCP networking
        let mut tcp_net = TcpNetwork::new(self.channel(), self.node_channel.clone(), self.current_api_version, self.current_value);
        self.tcp_channel.push(tcp_net.channel());
        let c = concurrency;
        let addr = self.tcp_address.clone();
        thread::spawn(move ||{
            tcp_net.start(c, addr.as_str());
        });

        match self.poll.register(&self.receiver_channel, RECEIVER_CHANNEL_TOKEN, Ready::readable(), PollOpt::edge()) {
            Ok(_) => {},
            Err(e) => {
                warn!("Unable to register channel receiver for base Networking. Need to Exit!! -> {}", e);
                process::exit(1);
            }
        }

        // making events for handling 5K events at once
        let mut events: Events = Events::with_capacity(5000);
        loop {
            let event_count = self.poll.poll(&mut events, None).unwrap();
            if event_count == 0 {
                continue;
            }

            for event in events.iter() {
                if event.token() == RECEIVER_CHANNEL_TOKEN {
                    // trying to get commands while there is available data
                    loop {
                        match self.receiver_channel.try_recv() {
                            Ok(cmd) => {
                                let mut c = cmd;
                                self.notify(&mut c);
                            }
                            // if we got error, then data is unavailable
                            // and breaking receive loop
                            Err(_) => break
                        }
                    }
                    continue;
                }
            }
        }
    }

    #[inline(always)]
    fn notify(&mut self, command: &mut NetworkCommand) {
        match command.cmd {
            NetworkCMD::HandleNewConnection => {
                if command.connection.len() == 0 {
                    return;
                }

                let conn = command.connection.remove(0);

                let mut e =  Event::default();
                e.name = String::from(EVENT_ON_CONNECTION_OPEN);
                e.from = conn.value;
                e.target = String::from("local");

                self.connections.insert(conn.value, conn);

                let _ = self.node_channel.send(NodeCommand {
                    cmd: NodeCMD::HandleDataEvent,
                    event: vec![e]
                });
            }
            NetworkCMD::HandleEventData => {
                if command.event.len() == 0 {
                    return;
                }

                let mut event = command.event.remove(0);
                let mut conns_to_send: Vec<&Connection> = Vec::new();
                for (value, conn) in &self.connections {
                    if event.path.div(*value) {
                        conns_to_send.push(conn);
                    }
                }

                let event_data = Arc::new(
                    match event.to_raw() {
                        Ok(d) => d,
                        Err(e) => {
                            warn!("Unable to convert event to bytes during write process -> {}", e);
                            return;
                        }
                    }
                );

                // we don't need event anymore
                drop(event);
                while !conns_to_send.is_empty() {
                    conns_to_send.remove(0).write_tcp(event_data.clone());
                }
            }
            NetworkCMD::TCPClientConnection => {
                if command.client_address.len() == 0 {
                    return;
                }

                self.tcp_connect(command.client_address.clone());
            }
        }
    }

    #[inline(always)]
    pub fn tcp_connect(&self, address: String) {
        if self.tcp_channel.len() != 1 {
            warn!("We don't have TCP networking service for making client connection");
            return;
        }

        let _ = self.tcp_channel[0].send(TcpNetworkCommand{
            cmd: TcpNetworkCMD::ClientConnection,
            client_address: address
        });
    }
}
