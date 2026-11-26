// 节点负载展示工具
// tunnelLimit: 隧道数量上限，0表示不限制
// tunnelCount: 已分配隧道数量（数据库统计）
// load: 节点实时上报负载 { clientCount, tunnelCount, connectionCount }

// 节点是否满载（隧道数量达到上限）
export function nodeFull(node) {
    return !!node && node.tunnelLimit > 0 && (node.tunnelCount || 0) >= node.tunnelLimit
}

// 隧道使用情况文本：x/y（未设置上限时只显示数量）
export function nodeTunnelText(node) {
    if (!node) {
        return '0'
    }
    const count = node.tunnelCount || 0
    if (node.tunnelLimit > 0) {
        return count + '/' + node.tunnelLimit
    }
    return '' + count
}

// 节点实时负载文本
export function nodeLoadText(node) {
    if (!node || node.online !== 1) {
        return '离线'
    }
    const load = node.load || {}
    return '客户端 ' + (load.clientCount || 0) + ' · 运行隧道 ' + (load.tunnelCount || 0) + ' · 连接 ' + (load.connectionCount || 0)
}
