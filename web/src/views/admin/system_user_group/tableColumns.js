import {h, ref} from "vue";
import {NButton, NSpace, NTag} from "naive-ui";

export function generateTableColumns(
    update = (row) => {
    },
    remove = (row) => {
    },
) {
    return [
        {title: '名称', key: 'name', ellipsis: {tooltip: true}, width: 150},
        {title: '排序', key: 'indexValue', width: 80},
        {
            title: '节点白名单',
            ellipsis: {tooltip: true},
            width: 120,
            render(row) {
                const count = row.nodeCodes ? row.nodeCodes.split(',').filter(v => v).length : 0
                return h(NTag, {size: 'small', type: count ? 'info' : 'success'}, {
                    default: () => count ? count + ' 个节点' : '全部节点'
                })
            }
        },
        {
            title: '套餐白名单',
            ellipsis: {tooltip: true},
            width: 120,
            render(row) {
                const count = row.configCodes ? row.configCodes.split(',').filter(v => v).length : 0
                return h(NTag, {size: 'small', type: count ? 'info' : 'success'}, {
                    default: () => count ? count + ' 个套餐' : '全部套餐'
                })
            }
        },
        {
            title: '隧道上限',
            width: 100,
            render(row) {
                return row.tunnelLimit > 0 ? row.tunnelLimit + ' 条' : '不限'
            }
        },
        {
            title: '月流量上限',
            width: 110,
            render(row) {
                return row.monthTraffic > 0 ? row.monthTraffic + ' GB' : '不限'
            }
        },
        {title: '用户数', key: 'userCount', width: 80},
        {title: '创建时间', key: 'createdAt', ellipsis: {tooltip: true}, width: 190},
        {
            title: '操作',
            key: 'operator',
            width: 134,
            render(row) {
                return generateButtonGroup(
                    generateUpdateButton(row, update),
                    generateDeleteButton(row, remove),
                )
            }
        },
    ]
}

function generateButtonGroup(...btn) {
    return h(NSpace, {}, {
        default: () => [...btn]
    })
}

function generateUpdateButton(row, event) {
    return h(NButton, {
        size: 'small',
        type: 'success',
        onClick() {
            event(row)
        },
    }, {
        default: () => '修改',
    })
}

function generateDeleteButton(row, event) {
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
                content: '确认删除该用户组吗?',
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
