import {h, ref} from "vue";
import {NButton, NSpace, NTag} from "naive-ui";

export function generateBatchTableColumns(
    viewCards = (row) => {
    },
    exportCards = (row) => {
    },
    remove = (row) => {
    },
) {
    return [
        {title: '批次名称', key: 'name', ellipsis: {tooltip: true}, width: 160},
        {
            title: '发放内容',
            ellipsis: {tooltip: true},
            width: 220,
            render(row) {
                const parts = []
                if (row.groupCode) {
                    const days = row.groupDays > 0 ? `${row.groupDays}天` : '永久'
                    parts.push(`${row.groupName || row.groupCode}（${days}）`)
                }
                if (parseFloat(row.amount) > 0) {
                    parts.push(`${row.amount} 积分`)
                }
                return parts.join(' + ') || '-'
            }
        },
        {
            title: '数量(已用/总数)',
            width: 130,
            render(row) {
                return `${row.usedCount} / ${row.total}`
            }
        },
        {title: '备注', key: 'remark', ellipsis: {tooltip: true}, width: 150},
        {title: '创建时间', key: 'createdAt', ellipsis: {tooltip: true}, width: 190},
        {
            title: '操作',
            key: 'operator',
            width: 220,
            render(row) {
                return generateButtonGroup(
                    generateButton(row, viewCards, '卡密', 'info'),
                    generateButton(row, exportCards, '导出', 'success'),
                    generateDeleteButton(row, remove, '确认删除该批次吗?已兑换的批次无法删除。'),
                )
            }
        },
    ]
}

export function generateCardTableColumns(
    ban = (row) => {
    },
) {
    return [
        {title: '卡密', key: 'secret', ellipsis: {tooltip: true}, width: 220},
        {
            title: '状态',
            width: 90,
            render(row) {
                const map = {
                    1: {label: '未使用', type: 'default'},
                    2: {label: '已兑换', type: 'success'},
                    3: {label: '已停用', type: 'error'},
                }
                const item = map[row.status] || {label: '未知', type: 'default'}
                return h(NTag, {size: 'small', type: item.type}, {default: () => item.label})
            }
        },
        {title: '兑换账号', key: 'account', ellipsis: {tooltip: true}, width: 130},
        {title: '兑换时间', key: 'usedAt', ellipsis: {tooltip: true}, width: 190},
        {title: '创建时间', key: 'createdAt', ellipsis: {tooltip: true}, width: 190},
        {
            title: '操作',
            key: 'operator',
            width: 100,
            render(row) {
                if (row.status === 2) {
                    return '-'
                }
                const toStatus = row.status === 3 ? 1 : 3
                const label = row.status === 3 ? '启用' : '停用'
                const type = row.status === 3 ? 'success' : 'error'
                return generateButton(row, (r) => ban(r, toStatus), label, type)
            }
        },
    ]
}

function generateButtonGroup(...btn) {
    return h(NSpace, {}, {
        default: () => [...btn]
    })
}

function generateButton(row, event, label, type) {
    return h(NButton, {
        size: 'small',
        type: type,
        onClick() {
            event(row)
        },
    }, {
        default: () => label,
    })
}

function generateDeleteButton(row, event, content) {
    return h(NButton, {
        size: 'small',
        type: 'error',
        onClick() {
            const loading = ref(false)
            $dialog.warning({
                maskClosable: true,
                closable: false,
                showIcon: true,
                title: '删除',
                content: content,
                negativeText: '取消',
                positiveText: '确认',
                negativeButtonProps: {
                    focusable: false,
                },
                loading: loading,
                positiveButtonProps: {
                    type: 'error',
                    focusable: false,
                },
                async onNegativeClick() {

                },
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
        default: () => '删除',
    })
}