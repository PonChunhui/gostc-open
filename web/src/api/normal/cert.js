import request from "../../request/index.js";

const baseUrl = '/v1/normal/cert'

export const apiNormalCertCreate = (data) => {
    return request.request({url: `${baseUrl}/create`, method: 'POST', data})
}

export const apiNormalCertUpdate = (data) => {
    return request.request({url: `${baseUrl}/update`, method: 'POST', data})
}

export const apiNormalCertDelete = (data) => {
    return request.request({url: `${baseUrl}/delete`, method: 'POST', data})
}

export const apiNormalCertPage = (data) => {
    return request.request({url: `${baseUrl}/page`, method: 'POST', data})
}

export const apiNormalCertIssue = (data) => {
    return request.request({url: `${baseUrl}/issue`, method: 'POST', data})
}

export const apiNormalCertRenew = (data) => {
    return request.request({url: `${baseUrl}/renew`, method: 'POST', data})
}

export const apiNormalCertUpload = (data) => {
    return request.request({url: `${baseUrl}/upload`, method: 'POST', data})
}
