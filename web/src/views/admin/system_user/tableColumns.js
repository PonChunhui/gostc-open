import {h, ref} from "vue";
import {NButton, NSpace, NTag, NTooltip} from "naive-ui";
import {flowFormat} from "../../../utils/flow.js";

export function generateTableColumns(
    update = (row) => {
    },
    remove = (row) => {
    },
    ban = (row) => {
    },
    unban = (row) => {
    },
) {
    return [
        {title: '账号', key: 'account', ellipsis: {tooltip: true}},
        {title: '邮箱', key: 'email', ellipsis: {tooltip: true}},
        {title: '备注', key: 'remark', ellipsis: {tooltip: true}},
        {
            title: '角色',
            key: 'admin',
            ellipsis: {tooltip: true},
            render(row) {
                return generateAdminTag(row)
            }
        },
        {
            title: '状态',
            key: 'status',
            width: 100,
            render(row) {
                return generateStatusTag(row)
            }
        },
        {title: '积分', key: 'amount', ellipsis: {tooltip: true}, width: 190},
        {
            title: '用户组',
            key: 'groupName',
            ellipsis: {tooltip: true},
            width: 130,
            render(row) {
                return h(NTag, {size: 'small', type: row.groupCode ? 'info' : ''}, {
                    default: () => row.groupName || '无组'
                })
            }
        },
        {
            title: '到期时间',
            key: 'groupExpAt',
            width: 190,
            render(row) {
                if (!row.groupExpAt) {
                    return h(NTag, {size: 'small'}, {
                        default: () => '永久'
                    })
                }
                if (row.prevGroupName) {
                    return h(NTooltip, {}, {
                        trigger: () => h('span', {}, row.groupExpAt),
                        default: () => `到期后回落：${row.prevGroupName}`
                    })
                }
                return row.groupExpAt
            }
        },
        {
            title: '流量(IN|OUT)',
            ellipsis: {tooltip: true},
            width: 190, render(row) {
                return flowFormat(row.inputBytes) + ' | ' + flowFormat(row.outputBytes);
            }
        },
        {title: '注册时间', key: 'createdAt', ellipsis: {tooltip: true}, width: 190},
        {
            title: '操作',
            key: 'operator',
            width: 210,
            render(row) {
                return generateButtonGroup(
                    generateUpdateButton(row, update),
                    generateBanButton(row, ban, unban),
                    generateDeleteButton(row, remove),
                )
            }
        },
    ]
}

function generateStatusTag(row) {
    if (row.status === 2) {
        return h(NTooltip, {}, {
            trigger: () => h(NTag, {size: 'small', type: 'error'}, {
                default: () => '已封禁'
            }),
            default: () => row.banReason || '已被管理员封禁'
        })
    }
    return h(NTag, {size: 'small', type: 'success'}, {
        default: () => '正常'
    })
}

function generateBanButton(row, ban, unban) {
    if (row.admin === 1) {
        return null
    }
    if (row.status === 2) {
        return h(NButton, {
            size: 'small',
            type: 'success',
            onClick() {
                unban(row)
            },
        }, {
            default: () => '解封',
        })
    }
    return h(NButton, {
        size: 'small',
        type: 'warning',
        onClick() {
            ban(row)
        },
    }, {
        default: () => '封禁',
    })
}

function generateAdminTag(row) {
    return h(NTag, {
        size: 'small',
        type: row.admin === 1 ? 'success' : '',
    }, {
        default: () => row.admin === 1 ? '管理员' : '普通用户'
    })
}

function generateButtonGroup(...btn) {
    return h(NSpace, {}, {
        default: () => btn.filter(b => b)
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
                content: '确认删除吗?',
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