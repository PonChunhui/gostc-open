import request from "../../request/index.js";

const baseUrl = '/v1/normal/pay'

export const apiNormalPayCreate = (data) => {
    return request.request({url: `${baseUrl}/create`, method: 'POST', data})
}

export const apiNormalPayQuery = (data) => {
    return request.request({url: `${baseUrl}/query`, method: 'POST', data})
}

export const apiNormalPayPage = (data) => {
    return request.request({url: `${baseUrl}/page`, method: 'POST', data})
}
