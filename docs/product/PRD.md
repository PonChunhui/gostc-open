# GOSTC 平台增强 · 产品需求文档（PRD）

> 版本：v1.0 ｜ 状态：待评审 ｜ 配套：[产品文档](产品文档.md)、[开发路线图](开发路线图.md)

## 通用约定

- **接口风格**：沿用现有 `POST /api/v1/{admin|normal|public|auth}/...`，请求头 `token: <jwt>`，响应 `{code, data, msg}`（code 0=成功）。新增接口在各模块列出。
- **数据模型**：gorm，sqlite/mysql 双驱动；新表/新列在 `server/bootstrap/persistence.go` 的 `autoMigrate` 注册；**只增不删**，旧数据兼容。
- **配置**：`server/configs.Config` 新增段，支持环境变量覆盖；每模块带 feature flag。
- **权限**：`admin`（管理员）/ `normal`（登录用户）/ `public`（公开）沿用现有中间件 `middleware.Auth` / `AuthAdmin`。
- **编号**：M1–M12 对应产品文档 12 项；每模块含 背景 / 用户故事 / 功能需求 / 接口与数据模型 / 验收标准 / 边界与失败处理。

---

## M1 隧道健康监控 + 告警 🔴

### 背景与痛点
客户端 ARPC 在线 ≠ 隧道可用。隧道可能未注册到 frps（如节点地址错误），或后端服务宕机，面板仍显示在线。需把「健康」拆成可观测的三态并主动探测，异常时告警。

### 用户故事
- 作为用户，我希望面板直接告诉我「隧道是否真的能用」，而不是只看到客户端在线。
- 作为用户，我希望后端服务挂了能立刻收到邮件/Webhook 通知，而不是等访问失败才发现。
- 作为管理员，我希望看到所有隧道的健康总览与历史。

### 功能需求
1. **健康三态**：
   - `client_online`：客户端 ARPC 控制通道在线（现有）。
   - `tunnel_registered`：代理已在 frps 注册（经 frps plugin / 节点查询确认）。
   - `backend_reachable`：主动探测后端真实可达。
2. **主动探测**（定时任务 `robfig/cron`，默认 30s，可配）：
   - host（HTTP/HTTPS）：向隧道入口发探测请求，判定非 frp404/非 5xx 为可达。
   - forward/tunnel/proxy（TCP）：对目标 `TargetIp:TargetPort` 经节点做 TCP dial。
   - p2p：探测 STCP 中继可达性（XTCP 直连不常驻，标注「按需」）。
3. **状态字段**：各隧道模型增 `HealthStatus`(0未知/1健康/2异常)、`LastCheckAt`、`LastError`。
4. **告警规则** `SystemAlertRule`：触发条件（隧道异常、节点离线、流量包耗尽、证书临期 N 天）、通知渠道（邮件/Webhook）、静默期、接收人。
5. **告警记录** `SystemAlertLog`：触发时间、对象、渠道、发送结果。
6. **面板**：隧道列表健康徽标 + 健康总览页 + 告警规则管理页。

### 接口与数据模型
- 新表 `SystemAlertRule{Code,Name,EventType,TargetKind,TargetCode,Channels,Receivers,SilenceSec,Enable}`、`SystemAlertLog{Code,RuleCode,EventType,ObjectKind,ObjectCode,Channel,Result,CreatedAt}`。
- 隧道模型增列 `HealthStatus int / LastCheckAt int64 / LastError string`。
- `POST /normal/gost/client/health/page`（健康总览）、`POST /normal/alert/rule/{create,update,delete,page}`、`POST /normal/alert/log/page`。
- 探测结果写缓存（Redis/local）+ 落库节流（状态变化才写）。

### 验收标准
- 后端服务停止后 ≤ 1 个探测周期内面板变红并触发告警。
- 节点地址错误导致未注册时，`tunnel_registered=false` 可见。
- 告警静默期内不重复轰炸；渠道发送失败有记录。

### 边界与失败处理
- 探测本身失败（节点不可达）≠ 后端故障，需区分标注，避免误报。
- 探测频率/并发可配，避免对后端造成压力；支持按隧道关闭探测。
- 告警渠道（邮件/Webhook）不可达时降级记录并重试，不阻塞探测。

---

## M2 安全加固 🔴

### 背景与痛点
密码 MD5+salt 弱哈希；密钥/凭据明文存储；无审计/账号级登录锁定；准入仅白名单。
> 现状校正：**TOTP 两步验证已实现**（`service.gen/open/close/login.otp.go`，`SystemUser.OtpKey`）；登录已有 **IP 安全 + 验证码**（`cache2.SetIpSecurity/ValidCaptcha`）。本期不重复造 2FA，仅补审计、账号级锁定、凭据加密、bcrypt、黑名单；2FA 可选增恢复码。

### 用户故事
- 作为用户，我希望密码安全存储、可开启两步验证。
- 作为管理员，我希望敏感操作有审计、暴力破解被拦、密钥不以明文落库。

### 功能需求
1. **bcrypt 密码**（`golang.org/x/crypto/bcrypt`）：登录成功时若旧哈希为 MD5 则透明 rehash 升级；新注册/改密直接用 bcrypt。
2. **凭据加密**：节点密钥、客户端密钥、VKey、DNS AK/Token 等用 AES-GCM 加密存储；主密钥来自配置/环境变量 `GOSTC_SECRET_KEY`；读取透明解密。
3. **2FA/TOTP**：**已存在**（`OtpKey` + GenOtp/OpenOtp/CloseOtp/LoginOtp）。本期仅可选增「恢复码」，不重复实现绑定/校验。
4. **审计日志** `SystemAuditLog`：gin 中间件记录登录/改密/删隧道/改节点/改配置等敏感操作（操作者、IP、对象、动作、结果、时间）。
5. **登录限流锁定**（`golang.org/x/time/rate`）：连续失败 N 次锁定 M 分钟；按账号+IP 限流。
6. **准入黑名单**：`GostClientAdmission` 增 `BlackEnable/BlackList`，与白名单并存（黑名单优先）。

### 接口与数据模型
- `SystemUser` 增 `TotpSecret string / TotpEnabled int / LoginFailCount int / LockUntil int64`。
- 新表 `SystemAuditLog{Code,UserCode,Action,ObjectKind,ObjectCode,Ip,Result,Detail,CreatedAt}`。
- `GostClientAdmission` 增 `BlackEnable int / BlackList string`。
- `POST /auth/2fa/{bind,verify,disable}`、`POST /admin/audit/page`、登录接口增 TOTP 校验分支。
- 加密工具 `server/pkg/crypto`（AES-GCM），密钥派生与版本前缀（便于轮换）。

### 验收标准
- 旧 MD5 用户无感登录且哈希升级为 bcrypt；库内无新增明文密钥。
- 开启 2FA 后无 TOTP 无法登录；恢复码可用一次。
- 连续失败触发锁定；审计可查询敏感操作。

### 边界与失败处理
- 主密钥缺失/错误：启动报错或解密失败明确提示，**绝不静默回退明文**。
- 加密迁移提供一次性脚本/启动迁移，幂等（已加密不重复加密，带版本前缀）。
- 2FA 丢失：恢复码 + 管理员重置流程。
- bcrypt 升级失败保留旧哈希仍可登录，不锁死用户。

---

## M3 Redis 缓存后端 🔴

### 背景与痛点
`server/pkg/memory/redis.impl.go` 与 `redis/redis.go` 全是 `panic("implement me")`；本地文件缓存无法多实例共享，阻碍水平扩展。

### 用户故事
- 作为管理员，我希望部署多个面板实例共享会话/缓存/在线状态，实现无状态水平扩展。

### 功能需求
1. 引入 `go-redis`，实现 `memory.Interface` 全部方法（SetStruct/GetStruct/SetString/GetString/Get/Set/Del/Sync）。
2. `bootstrap/memory.go` 按 `Config.Cache.Type`（local/redis）选择实现；redis 不可达启动告警并可降级 local。
3. 键命名空间前缀（多站点隔离）、TTL 正确、序列化兼容现有 `AnyType`。
4. 在线状态/会话/告警节流等共享态迁到 Redis。

### 接口与数据模型
- 配置段 `Cache{Type,Redis{Addr,Pwd,DB,Prefix}}`。
- 无新表；实现 `redis.impl.go`。

### 验收标准
- 两实例共享缓存：A 写入 B 可读；TTL 到期一致失效。
- 切回 local 行为不变；现有功能无回归。

### 边界与失败处理
- Redis 断连：降级 local + 告警，恢复后回切；不丢登录态（会话兜底）。
- 序列化版本兼容，旧缓存键可失效重建。

---

## M4 证书 DNS-01 自动化 🔴

### 背景与痛点
网关 `auto-set-cert` 仅 80 端口生效；非标端口/泛域名需手工 acme.sh。需面板化 ACME DNS-01 签发与续期。

### 用户故事
- 作为用户，我在面板填域名和 DNS 凭据，就能自动签出受信任证书并自动续期，无需登录服务器。

### 功能需求
1. 集成 ACME 库（`lego`），DNS-01 challenge；provider 插件 **DNSPod + Cloudflare + 阿里云 DNS**（已确认）。
2. `SystemCert` 模型：域名（含泛域名）、provider、凭据（加密）、证书/密钥路径、到期时间、自动续期开关、状态。
3. 定时续期（cron，到期前 30 天）；签发/续期结果下发到网关 proxy 与节点。
4. 支持上传自有证书；查看到期；手动触发签发。
5. DNS 凭据加密存储（复用 M2）。

### 接口与数据模型
- 新表 `SystemCert{Code,Domain,Provider,CredentialEnc,CertPath,KeyPath,ExpAt,AutoRenew,Status,LastError}`。
- `POST /normal/cert/{create,update,delete,page,issue,renew,upload}`。
- 配置段 `ACME{Email,CA,Provider默认}`。

### 验收标准
- 面板填域名+DNSPod 凭据 → 自动签出 Let's Encrypt 受信任证书（含泛域名）→ 网关加载 → 外网 HTTPS 无警告。
- 到期前自动续期并热加载。

### 边界与失败处理
- DNS 凭据错误/解析失败：明确报错、重试退避、不覆盖旧证书。
- CA 限流：退避与告警；签发失败保留旧证书继续服务。
- 凭据缺失/域名未解析：前置校验提示。

---

## M5 frp 高级能力表单化 🟡

### 背景与痛点
frp 引擎已支持负载均衡、健康检查、HTTP 路由、tcpmux、范围端口、带宽限制，但表单未暴露（`engine/common.go` 内嵌 `v1.*ProxyConfig` 已含这些字段）。

### 用户故事
- 作为高级用户，我希望在表单里配置多后端负载均衡、健康检查、按路径路由，而不必手写配置隧道。

### 功能需求
1. **负载均衡 + 健康检查**：host/forward 增 `LoadBalancer{Group,GroupKey}`、`HealthCheck{Type,Interval,Timeout,MaxFailed,Path}`，支持多后端。
2. **HTTP 路由**：host 增 `Locations []string`、`HostHeaderRewrite`、`RequestHeaders`。
3. **tcpmux**：forward 增 `Multiplexer`（httpconnect/tcpmux）、`RouteByHTTPUser`。
4. **范围端口**：forward 支持 `RemotePort` 范围（如 8000-8010）。
5. **带宽限制**：`ProxyTransport.BandwidthLimit/Mode` 表单化。

### 接口与数据模型
- `GostClientHost` 增 `Locations/LoadBalancerGroup/LoadBalancerGroupKey/HealthCheckType/HealthCheckPath/HealthCheckInterval/HealthCheckTimeout/HealthCheckMaxFailed/HostHeaderRewrite`。
- `GostClientForward` 增 `Multiplexer/RouteByHTTPUser/BandwidthLimit/BandwidthMode`。
- create/update 接口与 engine 映射（`arpc.client.engine.go`）扩展；naive-ui 表单增「高级」折叠区。

### 验收标准
- 表单配置两个后端 + 健康检查，停掉一个后端流量自动切到另一个。
- HTTP 按 location 路由生效；范围端口分配生效。

### 边界与失败处理
- 旧客户端不支持新字段：版本校验/降级提示。
- 健康检查参数非法（interval≤0）前置校验。
- 与现有套餐限速叠加时取更严格值，避免冲突。

---

## M6 网关（proxy）增强 🟡

### 背景与痛点
死配置 `sni/origin/disable-https/default` 未被读取；通配符域名匹配在非标准端口下失效（本地已修，待合入上游）；无 IP 路由；无 HTTP/3。

### 用户故事
- 作为管理员，我希望网关配置项都真实生效、通配符在任意端口可用、支持 IP 直接路由与 HTTP/3。

### 功能需求
1. **死配置清理**：移除或实现 `sni/origin/disable-https/default`（明确语义，文档同步）。
2. **通配符 bug 合入上游**：`proxy/pkg/proxy/proxy.go` Match() level2 用已剥端口的 host（本地已修），提交上游。
3. **IP 路由**：支持按 IP（无 Host/SNI）路由到指定后端。
4. **HTTP/3**：基于 `quic-go`（已在依赖）提供 HTTPS over QUIC 监听（可配开关）。
5. 配置热更新与回滚保留。

### 接口与数据模型
- proxy `config.yaml` schema 调整（清理死字段、增 `ip-routes`、`http3` 开关）。
- 无面板新表；网关配置经现有下发链路。

### 验收标准
- 通配符子域名在 1443 等非标端口正常匹配（回归本次修复）。
- IP 路由命中；HTTP/3 开关开启后 QUIC 可访问。
- 清理后无「写了不生效」的字段。

### 边界与失败处理
- HTTP/3 需 UDP 端口放行，失败降级 TCP；开关默认关。
- 配置变更热加载失败保留旧配置。

---

## M7 实时可观测性 🟡

### 背景与痛点
仅每日累计流量（`GostObs`），无实时连接数/速率/延迟，仪表盘单一。

### 用户故事
- 作为用户/管理员，我希望实时看到隧道的连接数、上下行速率、延迟与历史曲线。

### 功能需求
1. **指标采集**：frps plugin（newWorkConn/newUserConn/closeProxy）+ 客户端心跳上报连接数/字节/延迟。
2. **实时速率**：滑动窗口计算上下行速率；连接数 gauge。
3. **存储**：近期指标走 Redis（时序窗口），历史聚合落库（按分钟/小时）。
4. **仪表盘**：echarts 实时曲线、隧道/节点 TopN、健康总览（联动 M1）。

### 接口与数据模型
- 新表 `GostMetric{OriginKind,OriginCode,Metric,Value,Window,CreatedAt}`（聚合）。
- `POST /normal/metric/{realtime,history}`、dashboard 接口扩展。

### 验收标准
- 面板实时显示某隧道连接数与速率，误差可接受；历史曲线可查。

### 边界与失败处理
- 高频上报节流/聚合，避免压垮 DB；Redis 不可达降级内存窗口。
- 指标缺失（旧客户端）标注「不支持」。

---

## M8 节点高可用 + 调度 🟡

### 背景与痛点
节点单点，离线即隧道中断；无故障转移；DB 无读写分离（`dbresolver` 已在依赖未启用）。

### 用户故事
- 作为用户，我希望主节点离线时隧道自动切到备节点，不中断服务。
- 作为管理员，我希望数据库读写分离、节点按策略调度。

### 功能需求
1. **多节点注册**：客户端可同时注册到主/备节点。
2. **故障转移**：节点离线（M1 健康检查）→ 自动切换隧道到备节点并重推配置。
3. **节点分组与调度**：`GostNode` 增 `GroupCode/Priority/Weight`；创建隧道按策略选节点。
4. **读写分离**：启用 `gorm dbresolver`，配置主从。

### 接口与数据模型
- `GostNode` 增 `GroupCode/Priority/Weight`；隧道增 `BackupNodeCode/FailoverEnable`。
- 配置段 `DBResolver{Sources,Replicas}`。
- `POST /admin/gost/node/group/{create,page}`、隧道 update 增备节点。

### 验收标准
- 主节点宕机，隧道在探测周期内切到备节点并恢复访问。
- 读写分离生效（读走从库）。

### 边界与失败处理
- 无备节点/备节点也离线：明确告警，不静默失败。
- 切换抖动：加冷却期防反复切换；从库延迟监控。

---

## M9 支付与商业化 🟢

### 背景与痛点
仅积分/CDK，无在线支付与订单账单。

### 用户故事
- 作为用户，我希望在线充值/购买套餐；作为站长，我希望节点套餐可在线售卖、有账单。

### 功能需求
1. **可插拔支付网关**：默认支付宝 + 微信（需商户资质），Stripe 备选；统一 `PayProvider` 接口。
2. **订单/账单**：`SystemOrder{Code,UserCode,Amount,Provider,Status,TradeNo,...}`、`SystemBill`。
3. **回调验签**：异步通知验签、幂等、对账。
4. **充值→积分/套餐**：支付成功发放积分或开通套餐。
5. 沙箱模式（未配置商户时）。

### 接口与数据模型
- 新表 `SystemOrder/SystemBill`；配置段 `Pay{Providers,Alipay{},Wechat{},Stripe{}}`。
- `POST /normal/pay/{create,query}`、`POST /public/pay/notify/{provider}`（回调）。

### 验收标准
- 沙箱下单→回调→积分到账全链路；验签失败拒绝；重复回调幂等。

### 边界与失败处理
- 回调丢失：主动查单补偿；金额不一致拒绝；商户资质未配则仅沙箱。
- **已确认**：本期不做真实支付，仅做支付框架与沙箱；真实渠道（支付宝/微信/Stripe）后续接入。

---

## M10 OpenAPI + SDK + 批量/模板 🟢

### 背景与痛点
无接口文档/SDK，无批量操作与隧道模板，自动化与复用困难。

### 用户故事
- 作为开发者，我希望有 OpenAPI 文档与 SDK 用脚本管理隧道；作为用户，我希望批量启停/删除、用模板快速建隧道。

### 功能需求
1. **OpenAPI 文档**：从路由生成 swagger/openapi，在线浏览。
2. **API Token**：用户级长期 token（区别于登录 jwt），供 SDK/脚本。
3. **SDK**：Go/Python 封装常用接口。
4. **批量操作**：隧道批量启停/删除/迁移。
5. **隧道模板**：`SystemTunnelTemplate` 保存常用配置，一键套用。

### 接口与数据模型
- 新表 `SystemApiToken{Code,UserCode,TokenEnc,ExpAt}`、`SystemTunnelTemplate{Code,UserCode,Kind,Content}`。
- `POST /normal/api-token/{create,delete,page}`、`POST /normal/gost/client/batch`、`POST /normal/template/{create,page,apply}`。

### 验收标准
- OpenAPI 可浏览；API Token 鉴权可用；批量与模板生效。

### 边界与失败处理
- API Token 加密存储、可吊销、最小权限；批量部分失败返回逐项结果。

---

## M11 通知 / 自动更新 / 国际化 🟢

### 背景与痛点
通知仅邮件（M1 增 Webhook）；客户端无自动更新；前端无 i18n。

### 功能需求
1. **Webhook 通知**：通用 webhook 渠道（复用 M1 告警 + 事件订阅）。
2. **客户端自动更新**：版本检查接口 + 客户端自更新（校验签名/哈希，灰度）。
3. **i18n**：前端 naive-ui + 文案多语言（中/英），后端错误码文案。

### 接口与数据模型
- `POST /public/client/version`（最新版本/下载地址/哈希）；`SystemConfig` 增更新通道。
- 前端 `web/` 引入 vue-i18n，语言包。

### 验收标准
- 客户端检测到新版本可校验后自更新；前端可切换中英文。

### 边界与失败处理
- 自更新失败回滚旧版本；签名校验失败拒绝；i18n 缺省回退中文。

---

## M12 P2P 增强 🟢

### 背景与痛点
P2P 打洞成功率取决于 NAT，无预检；中继流量无独立计费。

### 功能需求
1. **NAT 类型预检**：访客连接前 STUN 探测 NAT 类型，提示直连成功率（EasyNAT 易/对称难）。
2. **中继计费**：STCP 中继流量单独统计与计费标识（联动套餐）。
3. **打洞策略优化**：KeepTunnelOpen、端口预测参数可配。

### 接口与数据模型
- `GostClientP2P` 增 `NatType/RelayBytes`；`POST /normal/gost/client/p2p/nat-check`。

### 验收标准
- 预检给出 NAT 类型与直连可能性提示；中继流量可统计。

### 边界与失败处理
- STUN 不可达：标注「未知」，不阻塞连接；对称 NAT 明确提示将走中继。

---

## 验收与测试总则

- 每模块关键路径补表驱动单测（重点：bcrypt 迁移、健康探测、ACME 签发、支付回调验签、Redis 实现）。
- `.cirrus.yml` 增 `go test ./...` + `web build`。
- 每阶段不得中断现有六类隧道；feature flag 可回滚。
- 详见 [开发路线图](开发路线图.md) 的阶段验收门禁。
