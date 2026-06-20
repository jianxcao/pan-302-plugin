use pan302_plugin_sdk::{
    encode_response, log, pb, read_proto,
};
use serde::Deserialize;
use serde_json::Value;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct StrmActionRequest {
    driver_id: String,
    cloud_path: String,
    #[serde(default)]
    task_name: String,
    #[serde(default)]
    force: bool,
    idempotency_key: String,
}

#[no_mangle]
pub unsafe extern "C" fn pan302_init(ptr: i32, len: i32) -> i64 {
    let req: pb::InitRequest = match read_proto(ptr, len) {
        Ok(value) => value,
        Err(error) => return encode_response(&pb::LifecycleResponse {
            ok: false,
            error: error.to_string(),
        }),
    };
    let plugin_name = req.plugin.as_ref().map(|p| p.name.as_str()).unwrap_or("");
    let plugin_ver = req.plugin.as_ref().map(|p| p.version.as_str()).unwrap_or("");
    log("info", &format!("STRM example plugin initialized: name={}, version={}", plugin_name, plugin_ver));
    encode_response(&pb::LifecycleResponse { ok: true, error: String::new() })
}

#[no_mangle]
pub unsafe extern "C" fn pan302_on_event(ptr: i32, len: i32) -> i64 {
    let event: pb::StrmEvent = match read_proto(ptr, len) {
        Ok(value) => value,
        Err(error) => return encode_response(&pb::LifecycleResponse {
            ok: false,
            error: error.to_string(),
        }),
    };
    log("info", &format!("received event {} ({})", event.event, event.event_id));
    encode_response(&pb::LifecycleResponse { ok: true, error: String::new() })
}

#[no_mangle]
pub unsafe extern "C" fn pan302_handle_http(ptr: i32, len: i32) -> i64 {
    let request: pb::HandleHttpRequest = match read_proto(ptr, len) {
        Ok(value) => value,
        Err(error) => return encode_response(&pb::HandleHttpResponse {
            status: 400,
            headers: std::collections::HashMap::new(),
            body: format!("error: {}", error).into_bytes(),
        }),
    };
    match request.path.as_str() {
        "/drivers" => match pan302_plugin_sdk::driver_list() {
            Ok(drivers) => {
                // 将 pb 消息或者其内部字段序列化为 JSON 字符串返回给接口
                // 我们可以直接把 pb 里的 drivers 转成 json 返回给 HTTP 客户端
                let mut driver_list = Vec::new();
                for d in drivers.drivers {
                    driver_list.push(serde_json::json!({
                        "id": d.id,
                        "name": d.name,
                        "type": d.r#type,
                        "typeGroup": d.type_group,
                        "health": d.health,
                        "capabilities": d.capabilities,
                    }));
                }
                let body = serde_json::to_vec(&driver_list).unwrap_or_default();
                encode_response(&pb::HandleHttpResponse {
                    status: 200,
                    headers: std::collections::HashMap::from([(
                        "content-type".to_string(),
                        "application/json; charset=utf-8".to_string()
                    )]),
                    body,
                })
            }
            Err(error) => {
                encode_response(&pb::HandleHttpResponse {
                    status: 502,
                    headers: std::collections::HashMap::new(),
                    body: format!("error: {}", error).into_bytes(),
                })
            }
        }
        "/write" | "/delete" => {
            let args: StrmActionRequest = match serde_json::from_slice(&request.body) {
                Ok(value) => value,
                Err(error) => return encode_response(&pb::HandleHttpResponse {
                    status: 400,
                    headers: std::collections::HashMap::new(),
                    body: format!("error: {}", error).into_bytes(),
                }),
            };
            let result = if request.path == "/write" {
                pan302_plugin_sdk::strm_write(pb::StrmWriteRequest {
                    driver_id: args.driver_id.clone(),
                    cloud_path: args.cloud_path.clone(),
                    task_name: args.task_name.clone(),
                    force: args.force,
                    idempotency_key: args.idempotency_key.clone(),
                })
            } else {
                pan302_plugin_sdk::strm_delete(pb::StrmDeleteRequest {
                    driver_id: args.driver_id.clone(),
                    cloud_path: args.cloud_path.clone(),
                    task_name: args.task_name.clone(),
                    idempotency_key: args.idempotency_key.clone(),
                })
            };
            match result {
                Ok(res) => {
                    let body = serde_json::to_vec(&serde_json::json!({
                        "status": res.status,
                        "localPath": res.local_path,
                        "cloudPath": res.cloud_path,
                    })).unwrap_or_default();
                    encode_response(&pb::HandleHttpResponse {
                        status: 200,
                        headers: std::collections::HashMap::from([(
                            "content-type".to_string(),
                            "application/json; charset=utf-8".to_string()
                          )]),
                        body,
                    })
                }
                Err(error) => {
                    encode_response(&pb::HandleHttpResponse {
                        status: 400,
                        headers: std::collections::HashMap::new(),
                        body: format!("error: {}", error).into_bytes(),
                    })
                }
            }
        }
        _ => {
            encode_response(&pb::HandleHttpResponse {
                status: 404,
                headers: std::collections::HashMap::new(),
                body: b"route not found".to_vec(),
            })
        }
    }
}

