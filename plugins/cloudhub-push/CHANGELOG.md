# Changelog

## 1.2.1

- 目录删除由宿主展开为文件级 `strm.deleted`；无 SHA1 的删除事件改为 DEBUG 跳过，避免正常空目录删除产生告警。

## 1.2.0

- 主程序从原始 STRM/media 事件派生独立的 `resource.ready`；CloudHub 新增统一消费资源就绪事件。
- STRM 来源资源由独立 Resource Readiness 模块延迟查询媒体详情，要求返回有效视频宽高才视为补全成功；原始 STRM 事件保持不变。
- 媒体详情最多查询 3 次，每次间隔 120 秒；第三次仍无分辨率时继续推送基础 STRM 信息。
- 主程序始终投递 STRM 与 Media 两种来源；`use_media_added_event` 只由 CloudHub 插件解释，不再影响宿主事件订阅。

## 1.1.0

- 删除只监听 `strm.deleted`，使用删除前文件快照中的 SHA1 清理 CloudHub owner。
- 新增默认监听 `strm.created` 并由宿主延迟 120 秒投递；可切换为 `media.item.added` 以补充媒体元数据。
- STRM 与 media 新增事件按配置互斥处理，避免重复推送。

## 1.0.2

- 推送资源时通过宿主媒体服务器配置调用 Emby/Jellyfin 接口，尝试补充视频、音频和容器字段；查询失败时继续推送基础资源。

## 1.0.1

- 增加 include 路径只有引入的路径才可以被上传,不设置默认为全部

## 1.0.0

- 初始化 CloudHub 资源推送功能，同步 STRM 创建、修改、移动与删除事件。
- 修复开启媒体服务代理时，配置页面静态资源加载 404 的问题。
- 修复保存设置后弹窗不自动关闭的问题。
