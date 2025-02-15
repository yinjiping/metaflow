mod consts;
pub mod decapsulate;
pub mod endpoint;
pub mod enums;
mod error;
pub mod flow;
pub mod lookup_key;
mod matched_field;
pub mod meta_packet;
pub mod platform_data;
pub mod policy;
pub mod port_range;
mod tag;
pub mod tagged_flow;
pub mod tap_port;
pub mod tap_types;

pub use consts::*;
pub use meta_packet::MetaPacket;
pub use platform_data::PlatformData;
pub use tagged_flow::TaggedFlow;
pub use tap_port::TapPort;
pub use tap_types::TapTyper;

use std::{
    fmt,
    hash::{Hash, Hasher},
    net::Ipv4Addr,
    sync::Arc,
};

use crate::proto::common::TridentType;

use policy::{Cidr, IpGroupData, PeerConnection};

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub struct XflowKey {
    ip: Ipv4Addr,
    tap_idx: u32,
}

impl Hash for XflowKey {
    fn hash<H: Hasher>(&self, state: &mut H) {
        let key = ((u32::from(self.ip) as u64) << 32) + self.tap_idx as u64;
        key.hash(state)
    }
}

impl fmt::Display for XflowKey {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "source_ip:{}, interface_index:{}", self.ip, self.tap_idx)
    }
}

pub trait FlowAclListener: Send + Sync {
    fn flow_acl_change(
        &mut self,
        trident_type: TridentType,
        ip_groups: &Vec<Arc<IpGroupData>>,
        platform_data: &Vec<Arc<PlatformData>>,
        peers: &Vec<Arc<PeerConnection>>,
        cidrs: &Vec<Arc<Cidr>>,
    );
    fn id(&self) -> usize;
}
