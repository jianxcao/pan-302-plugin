use crate::pb;
use crate::client::host_call_proto;

pub fn request(request: pb::HttpRequestArgs) -> Result<pb::HttpResponseData, String> {
    let resp = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::HttpRequest(request)),
    })?;
    match resp.result {
        Some(pb::host_response::Result::HttpResponse(r)) => Ok(r),
        _ => Err("expected HttpResponse response".to_string()),
    }
}
