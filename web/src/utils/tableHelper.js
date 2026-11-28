import {h, ref} from "vue";
import {NButton, NSpace} from "naive-ui";

// 通用表格操作按钮辅助（配合 vue-i18n 的 t 函数使用）。

export function generateButtonGroup(...btn) {
    return h(NSpace, {size: 'small'}, {
        default: () => [...btn]
    })
}

export function generateUpdateButton(row, event, t) {
    return h(NButton, {
        size: 'small',
        type: 'success',
        focusable: false,
        onClick() {
            event(row)
        },
    }, {
        default: () => t('common.update'),
    })
}

export function generateDeleteButton(row, event, t) {
    return h(NButton, {
        size: 'small',
        type: 'error',
        focusable: false,
        onClick() {
            const loading = ref(false)
            $dialog.warning({
                maskClosable: true,
                closable: false,
                showIcon: true,
                title: t('common.delete'),
                content: t('common.deleteConfirm'),
                negativeText: t('common.cancel'),
                positiveText: t('common.confirm'),
                negativeButtonProps: {focusable: false},
                loading: loading,
                positiveButtonProps: {type: 'error', focusable: false},
                async onPositiveClick() {
                    try {
                        loading.value = true
                        await event(row)
                    } finally {
                        loading.value = false
                    }
                }
            })
        },
    }, {
        default: () => t('common.delete'),
    })
}

export function generateTextButton(text, row, event, type = 'info') {
    return h(NButton, {
        size: 'small',
        type,
        focusable: false,
        onClick() {
            event(row)
        },
    }, {
        default: () => text,
    })
}
