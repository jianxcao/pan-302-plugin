# Host API 接口参考手册

Wasm 插件可以通过 SDK 发起 `host_call` 向宿主系统申请执行敏感操作。所有的 Host API 都被集中定义在 `HostRequest` 与 `HostResponse` 的 `oneof` 类型中，无需进行冗余的双重反序列化。

以下是目前支持的所有 Host API 说明。具体的结构体字段定义请直接查阅 `proto/plugin/v1/` 下对应的 `.proto` 文件。

---

## 1. Driver 操作 (driver.proto)

### driver.list (对应 `DriverListRequest`)
- **说明**：获取当前宿主系统中所有已挂载并处于启用状态的云网盘驱动列表。
- **返回**：`DriverListResponse` (包含所有网盘的 ID、名称、类型、运行健康状况和拥有的能力)。

### driver.read (对应 `DriverReadRequest`)
- **说明**：在指定的网盘驱动下获取指定路径/ID 的文件元数据，或者枚举目录。
- **参数**：
  - `driver_id` (String): 网盘 ID。
  - `action` (String): 支持 `"list"` (枚举目录) 或 `"file_info"` (获取单个文件元数据)。
  - `object` (ObjectRef): 对象标识（ID 或路径）。
  - `limit` (Int32): 分页大小限制。
- **返回**：`DriverReadResponse` (包含 `DriverObject` 的列表或单个对象，时间字段为 Timestamp 原生格式，Hashes map 中包含 SHA1/MD5 等)。

### driver.link (对应 `DriverLinkRequest`)
- **说明**：获取指定网盘文件对象的临时直链下载/播放地址。
- **返回**：`DriverLinkResponse` (包含生成的重定向直链 URL)。

### driver.mkdir / rename / delete / move / copy
- **说明**：基础的网盘文件系统写入、重命名、删除及跨目录复制移动操作。需要网盘本身具备对应能力权限（Cap）。

---

## 2. STRM 管理 (strm.proto)

### strm.write (对应 `StrmWriteRequest`)
- **说明**：触发宿主本地的文件系统生成对应的 `.strm` 文件，用于直接推流播放。
- **参数**：
  - `driver_id` (String): 源网盘 ID。
  - `cloud_path` (String): 云盘文件绝对路径。
  - `task_name` (String): 绑定的同步/生成任务名称。
  - `force` (Bool): 是否强制覆盖已有 strm。
  - `idempotency_key` (String): 幂等 Key，防止重复执行。
- **返回**：`StrmOperationResult` (包含写入的状态 `created`/`overwritten`，本地路径及云盘路径)。

### strm.delete (对应 `StrmDeleteRequest`)
- **说明**：删除本地生成的 `.strm` 文件。

---

## 3. 网络与系统能力

### http.request (对应 `HTTPRequestArgs` / `http.proto`)
- **说明**：通过宿主的 HTTP 客户端发起外部网络请求，以绕过 Wasm 沙箱无法直接使用宿主主机的网络套接字的限制。
- **参数**：
  - `method` (String)
  - `url` (String)
  - `headers` (map<string, StringList>)
  - `body` (Bytes): 直接以字节传输，杜绝 base64 效率问题。
- **返回**：`HTTPResponseData` (包含 `status` 状态码，`headers` 和字节类型的 `body`)。

### media.server_config.read (对应 `MediaServerConfigReadRequest` / `media.proto`)
- **权限**：`media.server_config.read`
- **说明**：读取宿主系统配置的 Emby/Jellyfin 连接信息。宿主只返回配置快照，不代理具体媒体业务 API；插件可结合 `http.request` 自行调用 `/Items` 或其他 Emby/Jellyfin 接口。
- **返回**：`MediaServerConfigReadResponse`，包含 `url`、`token`、`reverse_proxy_enabled`。未配置时返回空字符串字段。

### log.write (对应 `LogWriteRequest` / `log.proto`)
- **说明**：将插件自身的运行日志输出并汇入到宿主系统的结构化日志系统（由宿主 zap/loki 等管理）。

### config.read / config.write (config.proto)
- **说明**：读写插件在宿主数据库中存放的持久化自定义配置。配置类型是 `google.protobuf.Struct` 原生 JSON 对象。
