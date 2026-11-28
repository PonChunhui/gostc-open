$ErrorActionPreference = 'Stop'
$base = 'http://127.0.0.1:8080/api/v1'
$pass = 0; $fail = 0; $failures = @()
function Check($name, $cond) {
  if ($cond) { $script:pass++ } else { $script:fail++; $script:failures += $name; Write-Host "FAIL: $name" }
}
function Post($path, $tok, $body) {
  $h = @{}; if ($tok) { $h['token'] = $tok }
  $json = if ($body) { $body | ConvertTo-Json -Depth 8 -Compress } else { '{}' }
  return Invoke-RestMethod -Uri "$base$path" -Method Post -Headers $h -ContentType 'application/json' -Body $json
}
function Get2($path, $tok) {
  $h = @{}; if ($tok) { $h['token'] = $tok }
  return Invoke-RestMethod -Uri "$base$path" -Method Get -Headers $h
}

# ---- 登录 ----
$admin = (Post '/auth/login' $null @{account='admin';password='admin'})
Check 'admin.login' ($admin.code -eq 0)
$at = $admin.data.token
$alice = (Post '/auth/login' $null @{account='alice';password='123456'})
Check 'alice.login' ($alice.code -eq 0)
$ut = $alice.data.token

# ---- 管理员监控/管理 读 ----
Check 'admin.health.overview' ((Post '/admin/gost/client/health/overview' $at @{}).code -eq 0)
Check 'admin.health.page' ((Post '/admin/gost/client/health/page' $at @{page=1;size=5}).code -eq 0)
Check 'admin.alert.rule.page' ((Post '/admin/alert/rule/page' $at @{page=1;size=5}).code -eq 0)
Check 'admin.alert.log.page' ((Post '/admin/alert/log/page' $at @{page=1;size=5}).code -eq 0)
Check 'admin.metric.top' ((Post '/admin/metric/top' $at @{originKind=1;metric='input_rate';limit=5}).code -eq 0)
Check 'admin.node.page' ((Post '/admin/gost/node/page' $at @{page=1;size=5}).code -eq 0)
Check 'admin.node.group.page' ((Post '/admin/gost/node/group/page' $at @{page=1;size=5}).code -eq 0)
Check 'admin.client.version.page' ((Post '/admin/client/version/page' $at @{page=1;size=5}).code -eq 0)
Check 'admin.audit.page' ((Post '/admin/system/audit/page' $at @{page=1;size=5}).code -eq 0)
Check 'admin.user.page' ((Post '/admin/system/user/page' $at @{page=1;size=5}).code -eq 0)

# ---- 用户级 读 ----
Check 'normal.host.page' ((Post '/normal/gost/client/host/page' $ut @{page=1;size=10}).code -eq 0)
Check 'normal.forward.page' ((Post '/normal/gost/client/forward/page' $ut @{page=1;size=10}).code -eq 0)
Check 'normal.p2p.page' ((Post '/normal/gost/client/p2p/page' $ut @{page=1;size=10}).code -eq 0)
Check 'normal.client.page' ((Post '/normal/gost/client/page' $ut @{page=1;size=10}).code -eq 0)
Check 'normal.health.overview' ((Post '/normal/gost/client/health/overview' $ut @{}).code -eq 0)
Check 'normal.health.page' ((Post '/normal/gost/client/health/page' $ut @{page=1;size=10}).code -eq 0)
Check 'normal.alert.rule.page' ((Post '/normal/alert/rule/page' $ut @{page=1;size=5}).code -eq 0)
Check 'normal.alert.log.page' ((Post '/normal/alert/log/page' $ut @{page=1;size=5}).code -eq 0)
Check 'normal.cert.page' ((Post '/normal/cert/page' $ut @{page=1;size=5}).code -eq 0)
Check 'normal.pay.page' ((Post '/normal/pay/page' $ut @{page=1;size=5}).code -eq 0)
Check 'normal.template.page' ((Post '/normal/template/page' $ut @{page=1;size=5}).code -eq 0)

# ---- 指标 realtime/history ----
$top = (Post '/admin/metric/top' $at @{originKind=1;metric='input_rate';limit=5})
$nc = $top.data[0].code
Check 'metric.realtime' ((Post '/normal/metric/realtime' $ut @{originKind=1;originCode=$nc}).code -eq 0)
Check 'metric.history' ((Post '/normal/metric/history' $ut @{originKind=1;originCode=$nc;metric='input_rate'}).code -eq 0)

# ---- API Token 生命周期 ----
$tk = (Post '/normal/api-token/create' $ut @{name='reg-test';expDays=30})
Check 'apitoken.create' ($tk.code -eq 0)
$tkCode = $tk.data.code
Check 'apitoken.page' ((Post '/normal/api-token/page' $ut @{page=1;size=10}).code -eq 0)
if ($tkCode) { Check 'apitoken.delete' ((Post '/normal/api-token/delete' $ut @{code=$tkCode}).code -eq 0) }

# ---- 告警规则 生命周期 ----
$r = (Post '/normal/alert/rule/create' $ut @{name='reg-rule';eventType='tunnel_down';targetKind=1;channels=@('email');receivers=@('a@b.c');silenceSec=60;enable=1})
Check 'alert.rule.create' ($r.code -eq 0)
$rc = $r.data.code
if ($rc) {
  Check 'alert.rule.update' ((Post '/normal/alert/rule/update' $ut @{code=$rc;name='reg-rule2';eventType='tunnel_down';targetKind=1;channels=@('webhook');receivers=@('https://x');silenceSec=120;enable=1}).code -eq 0)
  Check 'alert.rule.delete' ((Post '/normal/alert/rule/delete' $ut @{code=$rc}).code -eq 0)
}

# ---- 模板 生命周期 ----
$tpl = (Post '/normal/template/create' $ut @{name='reg-tpl';kind='host';content=@{targetIp='192.168.1.10';targetPort='8080'}})
Check 'template.create' ($tpl.code -eq 0)
$tplCode = $tpl.data.code
if ($tplCode) { Check 'template.delete' ((Post '/normal/template/delete' $ut @{code=$tplCode}).code -eq 0) }

# ---- 证书 生命周期 ----
$c = (Post '/normal/cert/create' $ut @{domain='reg.demo.local';provider='cloudflare';credential=@{authToken='demo'};autoRenew=1})
Check 'cert.create' ($c.code -eq 0)
$cc = $c.data.code
if ($cc) { Check 'cert.delete' ((Post '/normal/cert/delete' $ut @{code=$cc}).code -eq 0) }

# ---- 支付 沙箱订单 ----
$o = (Post '/normal/pay/create' $ut @{amount='10';kind='recharge';provider='sandbox';subject='reg-test'})
Check 'pay.create' ($o.code -eq 0)
$oc = $o.data.code
if ($oc) { Check 'pay.query' ((Post '/normal/pay/query' $ut @{code=$oc}).code -eq 0) }

# ---- 隧道高级配置 + 启停（用演示数据，不依赖套餐）----
$hp = (Post '/normal/gost/client/host/page' $ut @{page=1;size=10})
$hostCode = $hp.data.list[0].code
if ($hostCode) {
  Check 'host.advanced' ((Post '/normal/gost/client/host/advanced' $ut @{code=$hostCode;locations=@('/api');hostHeaderRewrite='';healthCheckType='http';healthCheckPath='/';healthCheckInterval=30;healthCheckTimeout=5;healthCheckMaxFailed=3}).code -eq 0)
  Check 'host.disable' ((Post '/normal/gost/client/host/enable' $ut @{code=$hostCode;enable=2}).code -eq 0)
  Check 'host.enable' ((Post '/normal/gost/client/host/enable' $ut @{code=$hostCode;enable=1}).code -eq 0)
}
$fp = (Post '/normal/gost/client/forward/page' $ut @{page=1;size=10})
$fwdCode = $fp.data.list[0].code
if ($fwdCode) {
  Check 'forward.advanced' ((Post '/normal/gost/client/forward/advanced' $ut @{code=$fwdCode;multiplexer='tcpmux';bandwidthLimit='1MB';bandwidthMode='client';healthCheckType='tcp';healthCheckInterval=30}).code -eq 0)
}

# ---- 批量 ----
Check 'batch.enable' ((Post '/normal/gost/client/batch' $ut @{kind='host';action='enable';codes=@($hostCode)}).code -eq 0)

# ---- P2P NAT 预检 ----
$pp = (Post '/normal/gost/client/p2p/page' $ut @{page=1;size=10})
$p2pCode = $pp.data.list[0].code
if ($p2pCode) { Check 'p2p.natcheck' ((Post '/normal/gost/client/p2p/nat-check' $ut @{code=$p2pCode;natType='EasyNAT'}).code -eq 0) }

# ---- 公共接口 ----
Check 'public.openapi.spec' ((Get2 '/public/openapi/spec' $null).openapi -ne $null)
Check 'public.client.version' ((Get2 '/public/client/version/latest?platform=windows&arch=amd64&channel=stable' $null).code -eq 0)

Write-Host "==== SMOKE RESULT: pass=$pass fail=$fail ===="
if ($failures.Count) { Write-Host ("FAILURES: " + ($failures -join ', ')) }
