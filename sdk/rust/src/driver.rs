use crate::pb;
use crate::client::host_call_proto;

pub const CAP_LIST: &str = "list";
pub const CAP_LINK: &str = "link";
pub const CAP_MKDIR: &str = "mkdir";
pub const CAP_DELETE: &str = "delete";
pub const CAP_RENAME: &str = "rename";
pub const CAP_MOVE: &str = "move";
pub const CAP_COPY: &str = "copy";
pub const CAP_PUT: &str = "put";
pub const CAP_RAPID_UPLOAD: &str = "rapid_upload";
pub const CAP_STRM: &str = "strm";

pub fn list() -> Result<pb::DriverListResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverList(pb::DriverListRequest {})),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverList(r)) => Ok(r),
        _ => Err("expected DriverList response".to_string()),
    }
}

pub fn read(request: pb::DriverReadRequest) -> Result<pb::DriverReadResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverRead(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverRead(r)) => Ok(r),
        _ => Err("expected DriverRead response".to_string()),
    }
}

pub fn link(request: pb::DriverLinkRequest) -> Result<pb::DriverLinkResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverLink(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverLink(r)) => Ok(r),
        _ => Err("expected DriverLink response".to_string()),
    }
}

pub fn mkdir(request: pb::DriverMkdirRequest) -> Result<pb::DriverObject, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverMkdir(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverMkdir(r)) => Ok(r),
        _ => Err("expected DriverMkdir response".to_string()),
    }
}

pub fn rename(request: pb::DriverRenameRequest) -> Result<pb::DriverRenameResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverRename(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverRename(r)) => Ok(r),
        _ => Err("expected DriverRename response".to_string()),
    }
}

pub fn delete(request: pb::DriverDeleteRequest) -> Result<pb::DriverDeleteResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverDelete(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverDelete(r)) => Ok(r),
        _ => Err("expected DriverDelete response".to_string()),
    }
}

pub fn move_file(request: pb::DriverMoveCopyRequest) -> Result<pb::DriverMoveCopyResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverMove(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverMoveCopy(r)) => Ok(r),
        _ => Err("expected DriverMove response".to_string()),
    }
}

pub fn copy_file(request: pb::DriverMoveCopyRequest) -> Result<pb::DriverMoveCopyResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::DriverCopy(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::DriverMoveCopy(r)) => Ok(r),
        _ => Err("expected DriverCopy response".to_string()),
    }
}

pub fn has_capability(driver: &pb::DriverInfo, cap: &str) -> bool {
    driver.capabilities.iter().any(|c| c == cap)
}
