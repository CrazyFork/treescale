#![allow(dead_code)]
extern crate clap;

use network::NetworkConfig;
use helper::Log;
use std::error::Error;
use std::process;
use self::clap::{Arg, App};

pub struct NodeConfig {
    pub token: String,
    pub value: u64,
    pub network: NetworkConfig
}

pub struct MainConfig {
    pub node: NodeConfig,
    // TODO: here also could be resolver configurations
}

impl NodeConfig {
    /// Making default configurations
    pub fn default() -> NodeConfig {
        NodeConfig {
            token: String::new(),
            value: 0,
            network: NetworkConfig::default()
        }
    }
}

impl MainConfig {
    /// Getting default main configurations
    pub fn default() -> MainConfig {
        MainConfig {
            node: NodeConfig::default()
        }
    }

    /// Processing command line arguments and handling configurations
    pub fn process_cmd() -> MainConfig {
        let mut config = MainConfig::default();
        let matches = App::new("Treenity")
                        .about("TreeScale System Point Service, responsible for event distribution")
                        .arg(Arg::with_name("host")
                            .short("h")
                            .long("host")
                            .value_name("SERVER_HOST:PORT")
                            .help("Setting TCP Server Host:Port, default: 0.0.0.0:8000")
                            .takes_value(true))
                        .arg(Arg::with_name("token")
                            .short("t")
                            .long("token")
                            .value_name("SERVICE_NAME")
                            .help("Unique Name/Token for helping service discovery, if not provided, it would be set as a random hash string")
                            .takes_value(true))
                        .arg(Arg::with_name("concurrency")
                            .short("j")
                            .long("concurrency")
                            .value_name("NUM_CPU")
                            .help("Concurrency level for networking operations, default is all available cores")
                            .takes_value(true))
                        .get_matches();

        config.node.token = String::from(match matches.value_of("token") {
            Some(s) => s,
            None => ""
        });
        config.node.network.server_address = String::from(match matches.value_of("host") {
            Some(s) => s,
            None => "0.0.0.0:8000"
        });
        config.node.network.concurrency = match matches.value_of("concurrency") {
            Some(s) => {
                match s.parse::<usize>() {
                    Ok(n) => n,
                    Err(e) => {
                        Log::error("Concurrency Argument should valid positive number", e.description());
                        process::exit(1);
                    }
                }
            }
            None => 1
        };

        config
    }
}