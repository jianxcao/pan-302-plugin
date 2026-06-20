use crate::pb;
use crate::client::host_call_proto;

pub fn write(level: &str, message: &str, fields: std::collections::HashMap<String, String>) {
    let _ = host_call_proto(pb::HostRequest {
        id: String::new(),
        request: Some(pb::host_request::Request::LogWrite(pb::LogWriteRequest {
            level: level.to_string(),
            message: message.to_string(),
            fields,
        })),
    });
}

pub fn info(message: &str) {
    write("info", message, std::collections::HashMap::new());
}

pub fn warn(message: &str) {
    write("warn", message, std::collections::HashMap::new());
}

pub fn error(message: &str) {
    write("error", message, std::collections::HashMap::new());
}

pub mod notify {
    use crate::pb;
    use crate::client::host_call_proto;

    pub fn send(request: pb::NotifySendRequest) -> Result<pb::NotifySendResponse, String> {
        let resp = host_call_proto(pb::HostRequest {
            id: String::new(),
            request: Some(pb::host_request::Request::NotifySend(request)),
        })?;
        match resp.result {
            Some(pb::host_response::Result::NotifySend(r)) => Ok(r),
            _ => Err("expected NotifySend response".to_string()),
        }
    }
}
