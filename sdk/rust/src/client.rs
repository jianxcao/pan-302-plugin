use prost::Message;
use std::sync::atomic::{AtomicU64, Ordering};
use crate::pb;

static REQUEST_COUNTER: AtomicU64 = AtomicU64::new(1);

#[link(wasm_import_module = "pan302_v1")]
extern "C" {
    fn host_call(
        request_ptr: i32,
        request_len: i32,
        response_ptr: i32,
        response_capacity: i32,
    ) -> i32;
}

pub fn host_call_proto(request: pb::HostRequest) -> Result<pb::HostResponse, String> {
    let id = format!("rust-{}", REQUEST_COUNTER.fetch_add(1, Ordering::Relaxed));
    let mut req = request;
    req.id = id;

    let mut request_bytes = Vec::new();
    req.encode(&mut request_bytes).map_err(|e| e.to_string())?;

    let mut response_bytes = vec![0_u8; 4096];
    loop {
        let result = unsafe {
            host_call(
                request_bytes.as_ptr() as i32,
                request_bytes.len() as i32,
                response_bytes.as_mut_ptr() as i32,
                response_bytes.len() as i32,
            )
        };
        if result < 0 {
            return Err(format!("host_call failed with code {result}"));
        }
        let required = result as usize;
        if required > response_bytes.len() {
            response_bytes.resize(required, 0);
            continue;
        }
        response_bytes.truncate(required);
        let decoded = pb::HostResponse::decode(&response_bytes[..])
            .map_err(|error| error.to_string())?;

        if let Some(pb::host_response::Result::Error(err)) = decoded.result {
            return Err(format!("{}: {}", err.code, err.message));
        }
        return Ok(decoded);
    }
}
