# Proto 字段管理与 API 演进规范

为了保证插件宿主系统在后续升级中能够平滑地向前和向后兼容，所有 Host API 契约（Protobuf 结构）的修改必须严格遵守以下规范。

## 核心规则

### 1. 只新增、不删除、不重编号
所有的字段编号在分配后即与特定的字段语义绑定，**绝对不能修改、复用或删除**。
如果你需要弃用某个字段，不能直接将其删除，因为老版本的插件依然可能在读写该字段。正确做法是将其标记为 `reserved`：

```protobuf
message DriverObject {
  string id = 1;
  // ...
  reserved 15;          // 标记原 15 号字段已废弃，避免之后被重新分配
  reserved "old_field"; // 标记原字段名称已废弃
}
```

### 2. 所有新增字段必须默认可选 (optional)
在 proto3 中，基本类型字段默认是隐式可选的（如果未显式赋值则为零值），而 Message 类型也是引用可选的。
新增字段时，老插件因为不包含该字段，在反序列化时会安全地将其忽略；而新插件遇到未传该字段的老宿主时，也能优雅地通过默认零值回退。

### 3. 操作级版本控制 (Breaking Changes)
如果某项操作的改动导致语义发生根本性改变，且无法通过向前兼容的字段增量来实现，必须采用操作级版本升级：

1. **注册新操作名**：在 `HostRequest` / `HostResponse` 的 `oneof` 中分配一个新的字段与操作。例如，废弃 `driver.read`（对应 `DriverReadRequest`），改用 `driver.read.v2`（对应 `DriverReadV2Request`）。
2. **渐进式弃用**：宿主侧必须保留老操作的处理函数（例如 `driver.read` 的 handler）至少一个主要大版本周期，直到确保没有任何老插件再调用它。
3. **Manifest 权限更新**：新插件需要在 `manifest.json` 的 `permissions` 里显式声明调用 `driver.read.v2` 的权限。

---

## 编解码与交互最佳实践

- **零 base64 依赖**：对于二进制数据（如 HTTP Body），应直接使用 `bytes` 类型传输，避免使用 base64 等字符转码降低效率。
- **自描述 JSON 支持**：当插件需要交互复杂的多维动态配置时，请使用 `google.protobuf.Struct` 代替 `bytes` 以支持自描述序列化。
- **时间标准**：所有时间信息必须统一采用 `google.protobuf.Timestamp` 原生时间表示，不要使用自定义的 RFC3339 字符串。
