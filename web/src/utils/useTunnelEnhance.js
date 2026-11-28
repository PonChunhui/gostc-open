import {ref} from "vue";
import {apiNormalHealthPage} from "../api/normal/health.js";
import {apiNormalGostClientBatch} from "../api/normal/gost_client.js";

// 健康徽标：按隧道code映射健康状态
export function useHealth() {
    const healthMap = ref({})
    const loadHealth = async () => {
        try {
            const res = await apiNormalHealthPage({page: 1, size: 1000})
            const map = {}
            for (const item of (res.data.list || [])) {
                map[item.tunnelCode] = item
            }
            healthMap.value = map
        } catch (e) {
            // 健康接口不可用时静默降级
        }
    }
    // 返回 {type, text} 用于 n-tag
    const healthTag = (code) => {
        const item = healthMap.value[code]
        if (!item) return {type: 'default', text: '未知'}
        if (item.healthStatus === 1) return {type: 'success', text: '健康'}
        if (item.healthStatus === 2) return {type: 'error', text: '异常'}
        return {type: 'default', text: '未知'}
    }
    return {healthMap, loadHealth, healthTag}
}

// 批量操作：kind = host / forward / client
export function useBatch(kind) {
    const selected = ref([])
    const batchLoading = ref(false)
    const isSelected = (code) => selected.value.includes(code)
    const toggle = (code, checked) => {
        if (checked) {
            if (!selected.value.includes(code)) selected.value.push(code)
        } else {
            selected.value = selected.value.filter(c => c !== code)
        }
    }
    const clear = () => { selected.value = [] }
    const run = async (action, after) => {
        if (selected.value.length === 0) {
            $message.warning('请先选择记录')
            return
        }
        try {
            batchLoading.value = true
            const res = await apiNormalGostClientBatch({kind, action, codes: selected.value})
            const d = res.data || {}
            $message.success(`成功 ${d.success || 0} 项，失败 ${d.failed || 0} 项`)
            clear()
            if (after) after()
        } finally {
            batchLoading.value = false
        }
    }
    return {selected, batchLoading, isSelected, toggle, clear, run}
}
