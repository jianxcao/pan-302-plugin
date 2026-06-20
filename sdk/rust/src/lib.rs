use prost::Message;
use std::sync::atomic::{AtomicU64, Ordering};

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

pub mod pb {
    include!(concat!(env!("OUT_DIR"), "/plugin.v1.rs"));
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

pub fn encode_response<T: Message>(value: &T) -> i64 {
    let mut bytes = Vec::new();
    if value.encode(&mut bytes).is_err() {
        return 0;
    }
    let len = bytes.len();
    let mut boxed = bytes.into_boxed_slice();
    let ptr = boxed.as_mut_ptr() as u32;
    std::mem::forget(boxed);
    ((ptr as i64) << 32) | len as i64
}

pub unsafe fn read_proto<T: Message + Default>(
    ptr: i32,
    len: i32,
) -> Result<T, String> {
    if ptr < 0 || len < 0 {
        return Err("invalid pointer".to_string());
    }
    let bytes = std::slice::from_raw_parts(ptr as *const u8, len as usize);
    T::decode(bytes).map_err(|error| error.to_string())
}

#[no_mangle]
pub extern "C" fn pan302_alloc(size: i32) -> i32 {
    if size <= 0 {
        return 0;
    }
    let mut buffer = vec![0_u8; size as usize].into_boxed_slice();
    let ptr = buffer.as_mut_ptr();
    std::mem::forget(buffer);
    ptr as i32
}

#[no_mangle]
pub unsafe extern "C" fn pan302_free(ptr: i32, len: i32) {
    if ptr <= 0 || len <= 0 {
        return;
    }
    let slice = std::ptr::slice_from_raw_parts_mut(ptr as *mut u8, len as usize);
    drop(Box::from_raw(slice));
}
