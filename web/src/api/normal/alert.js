import request from "../../request/index.js";

const baseUrl = '/v1/normal/alert'

export const apiNormalAlertRuleCreate = (data) => {
    return request.request({url: `${baseUrl}/rule/create`, method: 'POST', data})
}

export const apiNormalAlertRuleUpdate = (data) => {
    return request.request({url: `${baseUrl}/rule/update`, method: 'POST', data})
}

export const apiNormalAlertRuleDelete = (data) => {
    return request.request({url: `${baseUrl}/rule/delete`, method: 'POST', data})
}

export const apiNormalAlertRulePage = (data) => {
    return request.request({url: `${baseUrl}/rule/page`, method: 'POST', data})
}

export const apiNormalAlertLogPage = (data) => {
    return request.request({url: `${baseUrl}/log/page`, method: 'POST', data})
}
