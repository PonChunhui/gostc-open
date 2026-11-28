# 入口跟随（Entry Follow）PRD

> P7 · 模块 E1–E4 · 2026-11-28 · 评审门禁：通过后才进入实现

## E1 DNS 跟随（host 自定义域名）
**背景**：自定义域名 DNS A 指旧节点 IP，转移后断。
**用户故事**：作为用户，我希望自定义域名在主节点宕机转移后自动解析到备节点，URL 不变、无需手动改 DNS。
**功能需求**：
- 转移成功后（host 且 GetCustomDomain 非空）自动把该域名 A 记录更新为备节点公网 IP。
- 支持 DNSPod / Cloudflare / 阿里云；TTL 可配（默认 60s）。
- 凭据优先复用该域名的 SystemCert（ACME），无则用全局 `configs.DNS`。
- 同时向备节点网关重推域名路由（NodeAddDomain）。
**接口/数据模型**：
- `server/service/common/dnsfollow.UpdateA(domain,ip string,ttl int)error`。
- `configs.DNS{Provider,Credential,TTL,Zone}`（config.go 增段，yaml `dns`）。
- zone 推导：`DNS.Zone` 优先，否则域名后两段。
**验收**：模拟主节点离线→转移→`dig 自定义域名` 返回备节点 IP；URL 可达。
**失败处理**：DNS 更新失败→告警+failover log 标记，但**不回滚**隧道切换（保隧道活着）；凭据缺失→跳过 DNS 仅告警。

## E2 同端口约束 + 全类型入口协调（含 p2p/tunnel/proxy）
**背景**：端口变化会让硬编码调用方失效；p2p/tunnel 访客 serverAddr 依赖节点地址。
**用户故事**：作为用户，我希望转移后公网端口不变；若备节点给不了同端口，宁可不要转移并告诉我。
**功能需求**：
- 定义各类公网端口（见产品文档表）。
- 转移前校验备节点能否提供同端口：forward/proxy 查 ForwardPorts 范围+占用（除自身）；tunnel/p2p/host 比对 bind/网关端口。
- **不满足→放弃转移**，发 `failover_port_conflict` 告警+冷却，写 log(reason=port_conflict)，不改 NodeCode。
- 满足→切换并计算 old/new Entry 交 E3。
**接口/数据模型**：
- `server/service/common/entryfollow`：`PublicPort(type,tunnel)int`、`CanServeSamePort(db,backup,type,tunnel)bool`、`Entry(type,node,tunnel)string`。
- 新告警事件 `model.ALERT_EVENT_FAILOVER_PORT_CONFLICT="failover_port_conflict"`。
**验收**：单测表驱动覆盖 forward 占用/越界、tunnel/p2p bind 不一致→拒绝；一致→通过。生产端口冲突时不切换且有告警。
**失败处理**：端口冲突=放弃+告警（不静默改端口）。

## E3 结构化 webhook + 入口变更历史
**背景**：调用方/运维需感知转移后的新入口。
**用户故事**：作为运维，我希望 webhook 收到结构化新旧入口，并能在面板查转移历史。
**功能需求**：
- `alert.DispatchPayload(...,payload map[string]any)`；webhook payload 合并 {event,objectKind,objectCode,tunnelType,oldEntry,newEntry,oldNode,newNode,portUnchanged}。旧 Dispatch 兼容。
- 新增 `GostFailoverLog{Base;TunnelCode/TunnelType/UserCode/FromNode/ToNode/OldEntry/NewEntry/Reason}`（reason: failover/failback/port_conflict）。
- 切换成功/冲突放弃均写 log；成功发 node_failover payload。
- API normal/admin `failover/log/page`；前端历史页（旧→新入口、端口不变徽标、reason）。
**验收**：webhook 接收端收到含 oldEntry/newEntry 的 JSON；历史页可见记录。
**失败处理**：webhook 失败记 SystemAlertLog(result2)，不影响转移。

## E4 可选 failback 回切
**背景**：主节点恢复后希望回切。
**用户故事**：作为运维，我希望主节点恢复后隧道自动回切（可开关）。
**功能需求**：`configs.Failover{FailbackEnable}`（默认关）；failoverScan 回切分支：读最近 GostFailoverLog，FromNode 在线且当前==ToNode 且非冷却且 CanServeSamePort→切回，写 reason=failback。
**验收**：开启后主节点恢复→回切+log；关闭则不回切。
**失败处理**：回切同样遵守同端口约束；不满足则保持现状。

## 横切
- AutoMigrate 只增不删；新 model 注册 generate+persistence。
- 不破坏现有五类隧道/failover/health；单测+CI 绿。
- DNS 库复用现有 indirect（cloudflare-go/dnspod-go/alibaba-cloud-sdk-go）提升为 direct。
