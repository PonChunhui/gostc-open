import {ref} from "vue";

// 表格多选 + 批量操作辅助（配合 n-data-table 的 row-key / checked-row-keys 使用）。

export function useTableBatch() {
    const checkedKeys = ref([])
    const onCheck = (keys) => {
        checkedKeys.value = keys || []
    }
    const clear = () => {
        checkedKeys.value = []
    }
    return {checkedKeys, onCheck, clear}
}

// 逐个调用现有单条 API 完成批量操作（不新增后端接口），汇总成功/失败并提示。
export function runBatch(t, keys, fn, done, opts = {}) {
    if (!keys || keys.length === 0) return
    const loading = ref(false)
    $dialog.warning({
        maskClosable: true,
        closable: false,
        showIcon: true,
        title: opts.title || t('common.batchDelete'),
        content: opts.content || t('common.batchDeleteConfirm', {n: keys.length}),
        negativeText: t('common.cancel'),
        positiveText: t('common.confirm'),
        negativeButtonProps: {focusable: false},
        loading: loading,
        positiveButtonProps: {type: opts.type || 'error', focusable: false},
        async onPositiveClick() {
            try {
                loading.value = true
                let success = 0, failed = 0
                for (const k of keys) {
                    try {
                        await fn(k)
                        success++
                    } catch (e) {
                        failed++
                    }
                }
                if (failed === 0) $message.success(t('common.batchResult', {success, failed}))
                else $message.warning(t('common.batchResult', {success, failed}))
                if (done) done()
            } finally {
                loading.value = false
            }
        }
    })
}
