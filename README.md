# pan-302-plugin

本仓库是 pan-302 插件系统的核心契约与 SDK 集合。它作为 git submodule 挂载在 pan-302 主仓库下，独立进行版本迭代与协议管理。

## 目录结构

- **`proto/plugin/v1/`** — 宿主与插件之间的所有交互行为（Wasm ABI v2）的 Protobuf 契约定义。
- **`sdk/`** — 支持不同语言编写 Wasm 插件的底层 SDK 开发工具：
  - `sdk/go/` — Go 语言版 SDK
  - `sdk/rust/` — Rust 语言版 SDK
- **`examples/`** — 示例插件实现，供开发参考：
  - `examples/strm-rust-example/` — 基于 Rust 编写的 STRM 示例插件
  - `examples/cloudhub-push/` — 基于 Go 编写的 CloudHub 推送插件
- **`docs/`** — 详细参考文档：
  - [快速开始指南](docs/getting-started.md)
  - [Host API 接口参考手册](docs/host-api-reference.md)
  - [Wasm ABI v2 二进制底层规范](docs/abi-spec.md)
  - [API 兼容性与字段演进规范](docs/api-evolution.md)

---

## 协议与代码生成

我们使用 `buf` 统一管理与校验所有的 Protobuf 文件。日常开发中，如果您修改了 `proto/` 目录下的 schema 定义，可以在根目录下执行以下命令：

```bash
# Lint 校验 proto 是否符合标准规范
buf lint

# 检查当前修改是否破坏了与先前版本的向后兼容性 (Breaking Change)
buf breaking --against '.git#branch=main'

# 生成对应的 Go pb 代码
buf generate
```
