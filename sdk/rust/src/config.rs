use crate::pb;
use crate::client::host_call_proto;

pub fn read() -> Result<pb::ConfigReadResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::ConfigRead(pb::ConfigReadRequest {})),
    })?;
    match resp.result {
        Some(pb::host_response::Result::ConfigRead(r)) => Ok(r),
        _ => Err("expected ConfigRead response".to_string()),
    }
}

pub fn write(request: pb::ConfigWriteRequest) -> Result<pb::ConfigWriteResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::ConfigWrite(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::ConfigWrite(r)) => Ok(r),
        _ => Err("expected ConfigWrite response".to_string()),
    }
}
