use crate::client::host_call_proto;
use crate::pb;

pub fn server_config() -> Result<pb::MediaServerConfigReadResponse, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::MediaServerConfigRead(
            pb::MediaServerConfigReadRequest {},
        )),
    })?;
    match resp.result {
        Some(pb::host_response::Result::MediaServerConfigRead(r)) => Ok(r),
        _ => Err("expected MediaServerConfigRead response".to_string()),
    }
}
