# 测试与本机验证

## 本机环境

当前开发机：

- 系统：Windows。
- NVIDIA GPU：NVIDIA GeForce RTX 5060 Ti。
- NVIDIA 驱动：591.74。
- `nvidia-smi`：可用。

已验证命令：

```powershell
nvidia-smi --query-gpu=name,uuid,driver_version,memory.total,memory.used,utilization.gpu,temperature.gpu,power.draw --format=csv,noheader,nounits
```

返回字段包含：

- GPU 型号。
- GPU UUID。
- 驱动版本。
- 总显存。
- 已用显存。
- GPU 利用率。
- 温度。
- 功耗。

文档中不记录完整 GPU UUID，避免把设备唯一标识写入仓库。

## 已完成的本机验证

### 构建和单元检查

```powershell
$env:GOCACHE='F:\project\GPUFleet\.gocache'
go test ./...
go build -o bin\gpufleet-server.exe .\cmd\gpufleet-server
go build -o bin\gpufleet-agent.exe .\cmd\gpufleet-agent
```

结果：

- `go test ./...` 通过。
- `gpufleet-server.exe` 构建成功。
- `gpufleet-agent.exe` 构建成功。

### 本机采集验证

```powershell
.\bin\gpufleet-agent.exe --print
```

结果包含：

- `NVIDIA GeForce RTX 5060 Ti`
- 驱动版本 `591.74`
- GPU 利用率。
- 显存占用。
- 温度。
- 功耗。
- 风扇、时钟、P-State、PCIe 链路字段。
- GPU UUID 以 SHA-256 哈希形式输出，不输出原始 UUID。

### 端到端验证

已在同一个 PowerShell 脚本中完成：

1. 启动 `gpufleet-server.exe` 到 `127.0.0.1:18080`。
2. 使用 `gpufleet-agent.exe -once` 上报本机 GPU 指标。
3. 登录 Web API。
4. 查询 `/api/v1/overview`。
5. 主动停止服务端进程。

验证结果：

```json
{
  "device_count": 1,
  "online_device_count": 1,
  "gpu_count": 1,
  "avg_util": 100,
  "first_gpu": "NVIDIA GeForce RTX 5060 Ti",
  "disk_status": "ok"
}
```

这证明当前 MVP 已经打通本机 Agent 采集、HMAC 上报、服务端接收、压缩存储、登录查询和 Web API 聚合。

## Agent 测试计划

### Windows

- 前台运行采集命令。
- 安装为 Windows Service。
- 验证服务重启后自动恢复。
- 验证网络断开时本地队列增长。
- 验证队列超过限制后丢弃旧样本。
- 验证服务端返回 `507` 时不无限重试同一批数据。

### Linux

- 前台运行采集命令。
- 安装为 systemd service。
- 验证无显示器环境。
- 验证驱动未加载时的错误状态。
- 验证多 GPU。
- 验证 MIG/ECC 字段在不支持设备上返回 null。

## 服务端测试计划

- HMAC 签名正确时接收。
- HMAC 签名错误时拒绝。
- 时间戳过期时拒绝。
- nonce 重复时拒绝。
- 请求体过大时拒绝。
- 磁盘低于 800MiB 时拒绝指标写入。
- gzip 请求体解压后大小限制。
- 当前 MVP 压缩分段文件保留清理。
- 后续 VictoriaMetrics 不可用时返回可诊断错误。
- 后续 SQLite 锁等待和恢复。

## 前端测试计划

- 桌面端 1440px。
- 平板端 768px。
- 手机端 390px。
- 浅色主题。
- 深色主题。
- 空数据状态。
- 设备离线状态。
- 磁盘保护状态。
- 图表密集数据状态。

## MVP 验收标准

- 一台 Windows 客户端可以上报 GPU 指标：已通过一次性上报验证。
- 服务端可以展示当前 GPU 状态：已通过 `/api/v1/overview` 验证。
- 服务端可以查询最近 1 小时历史曲线：API 已实现，仍需补 UI 验证。
- 设备断网上线状态正确变化：逻辑已实现，仍需补自动化验证。
- 服务端低磁盘空间时停止指标写入，并保留 800MiB 空闲空间：逻辑已实现，仍需补自动化验证。
- Web 面板在桌面和手机宽度下无明显布局错乱：响应式样式已实现，仍需浏览器截图验证。
