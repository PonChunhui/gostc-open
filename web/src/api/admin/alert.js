import request from "../../request/index.js";

const baseUrl = '/v1/admin/alert'

export const apiAdminAlertRulePage = (data) => {
    return request.request({url: `${baseUrl}/rule/page`, method: 'POST', data})
}

export const apiAdminAlertRuleDelete = (data) => {
    return request.request({url: `${baseUrl}/rule/delete`, method: 'POST', data})
}

export const apiAdminAlertLogPage = (data) => {
    return request.request({url: `${baseUrl}/log/page`, method: 'POST', data})
}
