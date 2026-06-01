# 当前实现说明

本文记录当前仓库中已经落地的实现，避免文档规划和代码状态脱节。

## 已实现

- Go module：`gpufleet`。
- 服务端命令：`cmd/gpufleet-server`。
- 客户端命令：`cmd/gpufleet-agent`。
- Agent 使用 `nvidia-smi` 只读采集 NVIDIA GPU 指标。
- Agent 支持 `--print` 本地采集调试模式。
- Agent 使用 HMAC-SHA256 签名主动上报。
- 服务端校验设备 ID、时间戳、nonce 和签名。
- 服务端拒绝重放 nonce。
- 服务端支持 gzip 请求体，并限制解压后的请求体大小。
- 服务端使用 gzip JSONL 分段文件保存时序指标。
- 服务端按保留期清理旧分段。
- 服务端默认保留 800MiB 磁盘空闲空间，低于阈值返回 `507`。
- 服务端使用 JSON 文件保存管理员、设备和审计元数据。
- Web 面板内置在服务端中，支持登录、总览、设备列表和 GPU 当前状态。

## 当前未实现

- Agent 本地离线队列。
- GPU 进程快照采集和展示。
- 统计报表接口和页面。
- 设备密钥轮换接口。
- Windows Service / systemd 安装脚本。
- VictoriaMetrics 存储适配。
- SQLite 元数据适配。
- 浏览器截图级 UI 验证。

## 运行边界

当前 MVP 不依赖外部数据库，适合先在单机公网服务端上验证链路。若设备数、保留时间或查询复杂度提高，应引入 VictoriaMetrics 作为时序后端。

服务端仍然不提供任何客户端控制能力。管理接口只影响服务端的设备记录和认证状态，不会修改客户端本地配置。

