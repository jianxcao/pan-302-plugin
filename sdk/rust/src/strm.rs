use crate::pb;
use crate::client::host_call_proto;

pub fn write(request: pb::StrmWriteRequest) -> Result<pb::StrmOperationResult, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::StrmWrite(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::StrmResult(r)) => Ok(r),
        _ => Err("expected StrmResult response".to_string()),
    }
}

pub fn delete(request: pb::StrmDeleteRequest) -> Result<pb::StrmOperationResult, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::StrmDelete(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::StrmResult(r)) => Ok(r),
        _ => Err("expected StrmResult response".to_string()),
    }
}
