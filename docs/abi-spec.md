# Wasm 宿主 ABI 规范 (ABI v2)

本规范描述了 pan-302 宿主系统与 Wasm 插件之间在二进制层面（ABI）的低级交互协议。

## 1. 核心导出函数 (Wasm -> Host)

每个符合 `pan302-plugin-abi/v2` 规范的 Wasm 模块都必须在根命名空间导出以下底层函数，以便宿主驱动生命周期和分发事件。

```typescript
// 1. 内存管理：由宿主调用，在 Wasm 内存中为即将传入的二进制数据开辟空间，返回内存首地址
function pan302_alloc(size: u32): u32;

// 2. 内存释放：宿主调用以释放已分配的内存块，避免内存泄漏
function pan302_free(ptr: u32, size: u32): void;

// 3. 插件初始化：宿主在实例化插件并载入初始配置后第一个调用的钩子
// 传入 InitRequest 消息在 Wasm 内存中的 ptr 和 len
// 返回一个 u64，其中高 32 位表示返回的 LifecycleResponse 消息的内存指针，低 32 位表示其长度
function pan302_init(ptr: u32, len: u32): u64;

// 4. 事件响应：宿主分发已订阅的 STRM 变更事件（如 strm.created, strm.deleted）
// 传入 StrmEvent 消息的 ptr 和 len，返回表示 LifecycleResponse 指针和长度的 u64
function pan302_on_event(ptr: u32, len: u32): u64;
```

### 可选导出函数

```typescript
// 5. 路由网关：若插件支持 UI 设置或自定义 HTTP API，需导出该函数以处理外部流入请求
// 传入 HandleHTTPRequest，返回表示 HandleHTTPResponse 消息指针和长度的 u64
function pan302_handle_http(ptr: u32, len: u32): u64;

// 6. 配置迁移：当插件版本升级且 manifest 中的 configVersion 升高时，宿主会调用此函数
// 传入 MigrateRequest，返回表示 MigrateResponse 消息指针和长度的 u64
function pan302_migrate(ptr: u32, len: u32): u64;
```

---

## 2. 宿主导入接口 (Host -> Wasm)

宿主环境只在命名空间 `"pan302_v1"` 中为 Wasm 提供一个系统接口：

```typescript
// 底层系统调用：发送 Protobuf 格式的 HostRequest，返回表示 HostResponse 的指针和长度的 packed u64
// Wasm 侧的 SDK 会在宿主调用该接口时分配内存，并在收到响应后反序列化为高级类型
function host_call(req_ptr: u32, req_len: u32, resp_ptr: u32, resp_cap: u32): i32;
```

---

## 3. 内存布局与数据流向

所有的数据传递全部基于 **Wasm 模块的线性内存**：
1. **数据传入（宿主 -> 插件）**：
   - 宿主调用 `pan302_alloc(len)`，在 Wasm 内存中申请一块大小为 `len` 的空间，得到首地址 `ptr`。
   - 宿主直接向 Wasm 模块的线性内存中 `ptr` 处写入序列化好的 Protobuf 数据。
   - 宿主调用 Wasm 导出的入口函数（如 `pan302_init(ptr, len)`）。
   - Wasm 内部函数处理完后，需要调用 `pan302_free(ptr, len)` 释放宿主申请的数据，或由 Wasm 逻辑自身在生命周期结束时手动清理。
2. **数据传出（插件 -> 宿主）**：
   - Wasm 内部序列化要返回的数据，保存在 Wasm 线性内存中。
   - 函数返回一个包含 `ptr` 和 `len` 的 `u64` 组合数（`packed = (ptr << 32) | len`）。
   - 宿主读取 Wasm 线性内存中 `[ptr, ptr + len]` 范围的数据并做反序列化。
   - 宿主处理完毕后，必须调用 Wasm 导出的 `pan302_free(ptr, len)` 释放在 Wasm 内存中产生的返回对象。
