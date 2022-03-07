use std::{
    collections::HashSet,
    fmt,
    io::ErrorKind,
    net::{IpAddr, Ipv4Addr, UdpSocket},
    time::{Duration, Instant},
};

use anyhow::{anyhow, Result};
use clap::{ArgEnum, Parser, Subcommand};

use trident::debug::{
    Beacon, Client, Message, Module, PlatformMessage, QueueMessage, RpcMessage, MAX_BUF_SIZE,
    SESSION_TIMEOUT,
};

const LISTENED_IP: &str = "::";
const LISTENED_PORT: u16 = 20035;
const ERR_PORT_MSG: &str = "error: The following required arguments were not provided:
    \t--port <PORT> required arguments were not provided";

#[derive(Parser)]
#[clap(name = "trident-ctl")]
struct Cmd {
    #[clap(subcommand)]
    command: ControllerCmd,
    /// remote trident listening port
    #[clap(short, long, parse(try_from_str))]
    port: Option<u16>,
    /// remote trident host ip, ipv6 format is 'fe80::5054:ff:fe95:c839', ipv4 format is '127.0.0.1'
    #[clap(short, long, parse(try_from_str), default_value_t=IpAddr::V4(Ipv4Addr::new(127, 0, 0, 1)))]
    address: IpAddr,
}

#[derive(Subcommand)]
enum ControllerCmd {
    Rpc(RpcCmd),
    Platform(PlatformCmd),
    Queue(QueueCmd),
    List,
}
#[derive(Parser)]
struct QueueCmd {
    /// monitor module, eg: trident-ctl queue --on xxxx --duration 40
    #[clap(long, validator = queue_name_validator, requires = "monitor")]
    on: Option<String>,
    /// monitor duration unit is second
    #[clap(long, group = "monitor")]
    duration: Option<u64>,
    /// turn off monitor, eg: trident-ctl queue --off xxx
    #[clap(long, validator = queue_name_validator)]
    off: Option<String>,
    /// show queue list, eg: trident-ctl queue --show
    #[clap(long)]
    show: bool,
    /// turn off all queue, eg: trident-ctl queue --clear
    #[clap(long)]
    clear: bool,
}

#[derive(Parser)]
struct PlatformCmd {
    /// Get resources with k8s api, eg: trident-ctl platform --k8s_get node
    #[clap(short, long, arg_enum)]
    k8s_get: Option<Resource>,
    /// Show k8s container mac to global interface index mappings, eg: trident-ctl platform --mac_mappings
    #[clap(short, long)]
    mac_mappings: bool,
}

#[derive(Clone, Copy, ArgEnum, Debug)]
enum Resource {
    Version,
    No,
    Node,
    Nodes,
    Ns,
    Namespace,
    Namespaces,
    Ing,
    Ingress,
    Ingresses,
    Svc,
    Service,
    Services,
    Deploy,
    Deployment,
    Deployments,
    Po,
    Pod,
    Pods,
    St,
    Statefulset,
    Statefulsets,
    Ds,
    Daemonset,
    Daemonsets,
    Rc,
    Replicationcontroller,
    Replicationcontrollers,
    Rs,
    Replicaset,
    Replicasets,
}

impl fmt::Display for Resource {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match *self {
            Resource::No | Resource::Node | Resource::Nodes => write!(f, "nodes"),
            Resource::Ns | Resource::Namespace | Resource::Namespaces => write!(f, "namespaces"),
            Resource::Svc | Resource::Service | Resource::Services => write!(f, "namespaces"),
            Resource::Deploy | Resource::Deployment | Resource::Deployments => {
                write!(f, "deployments")
            }
            Resource::Po | Resource::Pod | Resource::Pods => write!(f, "pods"),
            Resource::St | Resource::Statefulset | Resource::Statefulsets => {
                write!(f, "statefulsets")
            }
            Resource::Ds | Resource::Daemonset | Resource::Daemonsets => write!(f, "daemonsets"),
            Resource::Rc | Resource::Replicationcontroller | Resource::Replicationcontrollers => {
                write!(f, "replicationcontrollers")
            }
            Resource::Rs | Resource::Replicaset | Resource::Replicasets => {
                write!(f, "replicasets")
            }
            Resource::Ing | Resource::Ingress | Resource::Ingresses => write!(f, "ingresses"),
            Resource::Version => write!(f, "version"),
        }
    }
}

#[derive(Parser)]
struct RpcCmd {
    /// Get data from RPC, eg: trident-ctl rpc --get config
    #[clap(long, arg_enum)]
    get: RpcData,
}

#[derive(Clone, Copy, ArgEnum, Debug)]
enum RpcData {
    Config,
    Platform,
    TapTypes,
    Cidr,
    Groups,
    Acls,
    Segments,
    Version,
}

struct Controller {
    cmd: Option<Cmd>,
    addr: IpAddr,
    port: Option<u16>,
}

impl Controller {
    pub fn new() -> Self {
        let cmd = Cmd::parse();
        Self {
            addr: cmd.address,
            port: cmd.port,
            cmd: Some(cmd),
        }
    }

    fn dispatch(&mut self) -> Result<()> {
        match self.cmd.take().unwrap().command {
            ControllerCmd::Platform(c) => self.platform(c),
            ControllerCmd::Rpc(c) => self.rpc(c),
            ControllerCmd::List => self.list(),
            ControllerCmd::Queue(c) => self.queue(c),
        }
    }

    fn new_client() -> Result<Client> {
        let client = Client::new(Some(SESSION_TIMEOUT), (LISTENED_IP, 0))?;
        Ok(client)
    }

    /*
    $ trident-ctl list
    trident-ctl listening udp port 20035 to find trident

    -----------------------------------------------------------------------------------------------------
    VTAP ID        HOSTNAME                     IP                                            PORT
    -----------------------------------------------------------------------------------------------------
    1              ubuntu                       ::ffff:127.0.0.1                              42700
    */
    fn list(&self) -> Result<()> {
        let server = UdpSocket::bind((LISTENED_IP, LISTENED_PORT))?;
        server.set_read_timeout(Some(SESSION_TIMEOUT))?;
        let mut vtap_map = HashSet::new();

        println!(
            "trident-ctl listening udp port {} to find trident\n",
            LISTENED_PORT
        );
        println!("{:-<100}", "");
        println!(
            "{:<14} {:<28} {:45} {}",
            "VTAP ID", "HOSTNAME", "IP", "PORT"
        );
        println!("{:-<100}", "");
        loop {
            let mut buf = [0u8; 1024];
            let start = Instant::now();
            match server.recv_from(&mut buf) {
                Ok((n, a)) => {
                    if n == 0 {
                        continue;
                    }
                    let beacon: Beacon = bincode::deserialize(&buf[..n])?;

                    if !vtap_map.contains(&beacon.vtap_id) {
                        println!(
                            "{:<14} {:<28} {:<45} {}",
                            beacon.vtap_id,
                            beacon.hostname,
                            a.ip(),
                            a.port()
                        );
                        vtap_map.insert(beacon.vtap_id);
                    }
                }
                Err(e)
                    if start.elapsed() >= SESSION_TIMEOUT
                        && (cfg!(target_os = "windows") && e.kind() == ErrorKind::TimedOut
                            || cfg!(target_os = "linux") && e.kind() == ErrorKind::WouldBlock) =>
                {
                    // normal timeout, Window=TimedOut UNIX=WouldBlock
                    continue;
                }
                Err(e) => return Err(anyhow!("{}", e)),
            };
        }
    }

    fn rpc(&self, c: RpcCmd) -> Result<()> {
        if self.port.is_none() {
            return Err(anyhow!(ERR_PORT_MSG));
        }
        let client = Self::new_client()?;

        let payload = match c.get {
            RpcData::Acls => RpcMessage::Acls(None),
            RpcData::Config => RpcMessage::Config(None),
            RpcData::Platform => RpcMessage::PlatformData(None),
            RpcData::TapTypes => RpcMessage::TapTypes(None),
            RpcData::Cidr => RpcMessage::Cidr(None),
            RpcData::Groups => RpcMessage::Groups(None),
            RpcData::Segments => RpcMessage::Segments(None),
            RpcData::Version => RpcMessage::Version(None),
        };
        let msg = Message {
            module: Module::Rpc,
            msg: payload,
        };
        client.send_to(&msg, (self.addr, self.port.unwrap()))?;

        let mut buf = [0u8; MAX_BUF_SIZE];
        loop {
            let resp = client.recv::<Message<RpcMessage>>(&mut buf)?.into_inner();
            match resp {
                RpcMessage::Acls(v)
                | RpcMessage::PlatformData(v)
                | RpcMessage::TapTypes(v)
                | RpcMessage::Cidr(v)
                | RpcMessage::Groups(v)
                | RpcMessage::Segments(v) => match v {
                    Some(v) => {
                        for s in v {
                            println!("{}", s);
                        }
                    }
                    None => return Err(anyhow!(format!("cannot get {:?}", c.get))),
                },
                RpcMessage::Config(s) | RpcMessage::Version(s) => match s {
                    Some(s) => println!("{}", s),
                    None => return Err(anyhow!(format!("cannot get {:?}", c.get))),
                },
                RpcMessage::Fin => return Ok(()),
                RpcMessage::Err(e) => return Err(anyhow!(e)),
            }
            buf.fill(0);
        }
    }

    fn queue(&self, c: QueueCmd) -> Result<()> {
        if self.port.is_none() {
            return Err(anyhow!(ERR_PORT_MSG));
        }
        if c.on.is_some() && c.off.is_some() {
            return Err(anyhow!("error: --on and --off cannot set at the same time"));
        }

        let client = Self::new_client()?;
        if c.show {
            let msg = Message {
                module: Module::Queue,
                msg: QueueMessage::Names(None),
            };
            client.send_to(&msg, (self.addr, self.port.unwrap()))?;

            println!("available queues: ");

            let mut buf = [0u8; MAX_BUF_SIZE];
            loop {
                let res = client.recv::<Message<QueueMessage>>(&mut buf)?.into_inner();
                match res {
                    QueueMessage::Names(e) => match e {
                        Some(e) => {
                            for (i, s) in e.into_iter().enumerate() {
                                println!("{}. {}", i, s);
                            }
                        }
                        None => return Err(anyhow!("cannot get queue names")),
                    },
                    QueueMessage::Fin => return Ok(()),
                    QueueMessage::Err(e) => return Err(anyhow!(e)),
                    _ => unreachable!(),
                }
                // clear in case of dirty buffer
                buf.fill(0);
            }
        }

        if c.clear {
            let msg = Message {
                module: Module::Queue,
                msg: QueueMessage::Clear,
            };
            client.send_to(&msg, (self.addr, self.port.unwrap()))?;

            let mut buf = [0u8; MAX_BUF_SIZE];
            let res = client.recv::<Message<QueueMessage>>(&mut buf)?.into_inner();
            match res {
                QueueMessage::Fin => return Ok(()),
                QueueMessage::Err(e) => return Err(anyhow!(e)),
                _ => unreachable!(),
            }
        }

        if let Some(s) = c.off {
            let msg = Message {
                module: Module::Queue,
                msg: QueueMessage::Off(s),
            };
            client.send_to(&msg, (self.addr, self.port.unwrap()))?;

            let mut buf = [0u8; MAX_BUF_SIZE];
            let res = client.recv::<Message<QueueMessage>>(&mut buf)?.into_inner();
            match res {
                QueueMessage::Fin => return Ok(()),
                QueueMessage::Err(e) => return Err(anyhow!(e)),
                _ => unreachable!(),
            }
        }

        if let Some((s, d)) = c.on.zip(c.duration) {
            if d == 0 {
                return Err(anyhow!("zero duration isn't allowed"));
            }

            let dur = Duration::from_secs(d);

            let msg = Message {
                module: Module::Queue,
                msg: QueueMessage::On((s, dur)),
            };
            client.send_to(&msg, (self.addr, self.port.unwrap()))?;

            let mut buf = [0u8; MAX_BUF_SIZE];
            let res = client.recv::<Message<QueueMessage>>(&mut buf)?.into_inner();
            if let QueueMessage::Err(e) = res {
                return Err(anyhow!(e));
            }
            // clear in case of dirty buffer
            buf.fill(0);
            let mut msg_seq = 0;
            loop {
                let res = client.recv::<Message<QueueMessage>>(&mut buf)?.into_inner();
                match res {
                    QueueMessage::Send(e) => {
                        for s in e {
                            println!("MSG-{} {}", msg_seq, s);
                            msg_seq += 1;
                        }
                    }
                    QueueMessage::Fin => return Ok(()),
                    QueueMessage::Err(e) => return Err(anyhow!(e)),
                    _ => unreachable!(),
                }
                // clear in case of dirty buffer
                buf.fill(0);
            }
        }

        Ok(())
    }

    fn platform(&self, c: PlatformCmd) -> Result<()> {
        if self.port.is_none() {
            return Err(anyhow!(ERR_PORT_MSG));
        }
        let client = Self::new_client()?;
        if c.mac_mappings {
            let msg = Message {
                module: Module::Platform,
                msg: PlatformMessage::MacMappings(None),
            };
            client.send_to(&msg, (self.addr, self.port.unwrap()))?;
            println!("Interface Index \t MAC address");

            let mut buf = [0u8; MAX_BUF_SIZE];
            loop {
                let res = client
                    .recv::<Message<PlatformMessage>>(&mut buf)?
                    .into_inner();
                match res {
                    PlatformMessage::MacMappings(e) => {
                        match e {
                            /*
                            $ trident-ctl -p 42700 platform --mac-mappings
                            Interface Index          MAC address
                            12                       01:02:03:04:05:06
                            13                       01:02:03:04:05:06
                            14                       01:02:03:04:05:06
                            */
                            Some(e) => {
                                for (idx, m) in e {
                                    println!("{:<15} \t {}", idx, m);
                                }
                            }
                            None => return Err(anyhow!("cannot get mac mappings")),
                        }
                    }
                    PlatformMessage::Fin => return Ok(()),
                    _ => unreachable!(),
                }
                // clear in case of dirty buffer
                buf.fill(0);
            }
        }

        if let Some(r) = c.k8s_get {
            if let Resource::Version = r {
                let msg = Message {
                    module: Module::Platform,
                    msg: PlatformMessage::Version(None),
                };
                client.send_to(&msg, (self.addr, self.port.unwrap()))?;
                let mut buf = [0u8; MAX_BUF_SIZE];
                loop {
                    let res = client
                        .recv::<Message<PlatformMessage>>(&mut buf)?
                        .into_inner();
                    match res {
                        PlatformMessage::Version(v) => {
                            /*
                            $ trident-ctl -p 54911 platform --k8s-get version
                            k8s-api-watcher-version xxx
                            */
                            match v {
                                Some(v) => println!("{}", v),
                                None => return Err(anyhow!("cannot get server version")),
                            }
                        }
                        PlatformMessage::Fin => return Ok(()),
                        _ => unreachable!(),
                    }
                    // clear in case of dirty buffer
                    buf.fill(0);
                }
            }

            let msg = Message {
                module: Module::Platform,
                msg: PlatformMessage::WatcherReq(r.to_string()),
            };

            client.send_to(&msg, (self.addr, self.port.unwrap()))?;
            let mut buf = [0u8; MAX_BUF_SIZE];
            loop {
                let res = client
                    .recv::<Message<PlatformMessage>>(&mut buf)?
                    .into_inner();
                match res {
                    PlatformMessage::WatcherRes(v) => {
                        /*
                        $ trident-ctl -p 54911 platform --k8s-get node
                        nodes entries...
                        */
                        match v {
                            Some(v) => {
                                for e in v {
                                    println!("{}", e);
                                }
                            }
                            None => return Err(anyhow!("cannot get watcher entries")),
                        }
                    }
                    PlatformMessage::Fin => return Ok(()),
                    _ => unreachable!(),
                }
                // clear in case of dirty buffer
                buf.fill(0);
            }
        }
        Ok(())
    }
}

fn queue_name_validator(s: &str) -> Result<(), &'static str> {
    match s {
        "1-tagged-flow-to-quadruple-generator"
        | "1-packet-statistics-to-doc"
        | "1-mini-meta-packet-to-pcap"
        | "2-flow-with-meter-to-second-collector"
        | "2-flow-with-meter-to-minute-collector"
        | "2-second-flow-to-minute-aggrer"
        | "2-doc-to-collector-sender"
        | "3-flow-to-collector-sender" => Ok(()),
        _ => Err("invalid queue name"),
    }
}

fn main() {
    let mut controller = Controller::new();
    if let Err(e) = controller.dispatch() {
        eprintln!("{}", e);
    }
}
