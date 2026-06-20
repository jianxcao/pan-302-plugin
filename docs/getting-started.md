# 插件开发快速开始

本指南将帮助您快速了解如何安装开发工具、创建一个简单的 pan-302 插件，并编译打包它。

## 环境准备

### 1. 安装 Wasm/Protobuf 编译工具

#### Go 开发者
- **Go 1.25.1+** (支持 `wasip1` 编译架构)
- **buf**: Proto 生成管理工具：
  ```bash
  brew install bufbuild/buf/buf
  ```
- **protoc-gen-go**: Go Protobuf 代码生成器：
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  ```

#### Rust 开发者
- **Rust toolchain**: 建议使用 `rustup` 安装最新 stable 版本。
- **wasm32-unknown-unknown target**: 
  ```bash
  rustup target add wasm32-unknown-unknown
  ```
- **protobuf** (包含 `protoc` 编译器): 用于 `prost-build` 在编译期编译 proto 契约：
  ```bash
  brew install protobuf
  ```

---

## 创建并构建您的第一个插件

本项目在对应文件夹下提供了现成的脚手架和参考插件：
- `examples/strm-rust-example`: 基于 Rust SDK 编写的极简示例插件（含简单前端 UI 演示）。
- `examples/strm-go-example`: 基于 Go SDK 编写的极简 Go 语言示例插件（含免编译的原生 JavaScript 配置界面）。
- `plugins/cloudhub-push` [官方插件]: 基于 Go SDK 编写的 CloudHub 资源推送官方正式插件，采用了现代化的 **Vite 编译 + Vue SFC (单文件组件)** 架构。

下面以 Rust 插件开发为例介绍开发流程：

### 1. 修改 Manifest 配置文件
每个插件都必须在根目录下包含一个 `manifest.json`。
```json
{
  "schemaVersion": 1,
  "name": "my-cool-plugin",
  "version": "1.0.0",
  "displayName": "我的酷炫插件",
  "abi": "pan302-plugin-abi/v2",
  "wasm": "plugin.wasm",
  "permissions": ["log.write", "driver.list"]
}
```

### 2. 编写插件代码
在 `src/lib.rs` 中，您可以利用 `pan302-plugin-sdk` 调用宿主函数：

```rust
use pan302_plugin_sdk::{pb, log, encode_response, read_proto};

#[no_mangle]
pub unsafe extern "C" fn pan302_init(ptr: i32, len: i32) -> i64 {
    log("info", "Hello from Wasm plugin!");
    encode_response(&pb::LifecycleResponse { ok: true, error: String::new() })
}
```

### 3. 编译 WASM 产物
在插件目录运行：
```bash
cargo build --target wasm32-unknown-unknown --release
```
编译产物将会保存在 `target/wasm32-unknown-unknown/release/` 下。

### 4. 插件打包
编译出 `.wasm` 文件后，与 `manifest.json`、`default-config.json` 等一起放置在打包目录下，使用宿主自带的 `pan302-plugin` 命令行工具进行打包：
```bash
pan302-plugin build ./package-dir --output ./my-plugin.panplugin
```

打包出来的 `.panplugin` 文件即可在 pan-302 后台直接上传并激活。

---

## 插件前端 UI 开发

如果您需要为插件开发自定义的控制或配置面板，并希望与宿主**共用同一个 Vue 与 Naive UI 依赖以大幅压缩打包体积并自适应跟随主题**，请务必阅读：
👉 [插件前端 UI 扩展开发指南](ui-development-guide.md)

