import request from "../request/index.js";

const baseUrl = '/config'

export const apiConfigQuery = (data) => {
    return request.request({
        url: `${baseUrl}/query`, method: 'POST', data
    })
}

export const apiConfigUpdateQuery = (data) => {
    return request.request({
        url: `${baseUrl}/update/query`, method: 'POST', data
    })
}

export const apiConfigUpdateSave = (data) => {
    return request.request({
        url: `${baseUrl}/update/save`, method: 'POST', data
    })
}

export const apiConfigUpdateCheck = (data) => {
    return request.request({
        url: `${baseUrl}/update/check`, method: 'POST', data
    })
}

export const apiConfigUpdateApply = (data) => {
    return request.request({
        url: `${baseUrl}/update/apply`, method: 'POST', data
    })
}
