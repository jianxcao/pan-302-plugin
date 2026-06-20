pub mod pb {
    include!(concat!(env!("OUT_DIR"), "/plugin.v1.rs"));
}

pub mod abi;
pub mod client;

pub mod driver;
pub mod strm;
pub mod config;
pub mod http;
pub mod logger;

pub use abi::*;
pub use client::*;

// 向后兼容平铺命名空间导出 (Deprecated aliases for backward compatibility)
pub use driver::list as driver_list;
pub use driver::read as driver_read;
pub use driver::link as driver_link;
pub use driver::mkdir as driver_mkdir;
pub use driver::rename as driver_rename;
pub use driver::delete as driver_delete;
pub use driver::move_file as driver_move;
pub use driver::copy_file as driver_copy;
pub use driver::has_capability;

pub use strm::write as strm_write;
pub use strm::delete as strm_delete;

pub use config::read as config_read;
pub use config::write as config_write;

pub use http::request as http_request;

pub use logger::notify::send as notify_send;

#[deprecated(since = "0.2.0", note = "use logger::write or logger::info instead")]
pub fn log(level: &str, message: &str) {
    logger::write(level, message, std::collections::HashMap::new());
}

