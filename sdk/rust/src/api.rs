use crate::pb;
use crate::client::host_call_proto;

// 插件所支持的网盘能力位预设常量
pub mod cap {
    pub const LIST: &str = "list";
    pub const LINK: &str = "link";
    pub const MKDIR: &str = "mkdir";
    pub const DELETE: &str = "delete";
    pub const RENAME: &str = "rename";
    pub const MOVE: &str = "move";
    pub const COPY: &str = "copy";
    pub const PUT: &str = "put";
    pub const RAPID_UPLOAD: &str = "rapid_upload";
    pub const STRM: &str = "strm";
}

pub fn log(level: &str, message: &str) {
    let _ = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::LogWrite(pb::LogWriteRequest {
            level: level.to_string(),
            message: message.to_string(),
            fields: std::collections::HashMap::new(),
        })),
    });
}

pub fn config_read() -> Result<pb::ConfigReadResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::ConfigRead(pb::ConfigReadRequest {})),
    })?;
    match resp.result {
        Some(pb::host_response::Result::ConfigRead(r)) => Ok(r),
        _ => Err("expected ConfigRead response".to_string()),
    }
}

pub fn config_write(request: pb::ConfigWriteRequest) -> Result<pb::ConfigWriteResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::ConfigWrite(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::ConfigWrite(r)) => Ok(r),
        _ => Err("expected ConfigWrite response".to_string()),
    }
}

pub fn driver_list() -> Result<pb::DriverListResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverList(pb::DriverListRequest {})),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverList(r)) => Ok(r),
        _ => Err("expected DriverList response".to_string()),
    }
}

pub fn strm_write(request: pb::StrmWriteRequest) -> Result<pb::StrmOperationResult, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::StrmWrite(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::StrmResult(r)) => Ok(r),
        _ => Err("expected StrmResult response".to_string()),
    }
}

pub fn strm_delete(request: pb::StrmDeleteRequest) -> Result<pb::StrmOperationResult, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::StrmDelete(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::StrmResult(r)) => Ok(r),
        _ => Err("expected StrmResult response".to_string()),
    }
}

pub fn notify_send(request: pb::NotifySendRequest) -> Result<pb::NotifySendResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::NotifySend(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::NotifySend(r)) => Ok(r),
        _ => Err("expected NotifySend response".to_string()),
    }
}

pub fn has_capability(driver: &pb::DriverInfo, cap: &str) -> bool {
    driver.capabilities.iter().any(|c| c == cap)
}
