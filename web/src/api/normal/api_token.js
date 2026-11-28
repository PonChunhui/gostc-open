import request from "../../request/index.js";

const baseUrl = '/v1/normal/api-token'

export const apiNormalApiTokenCreate = (data) => {
    return request.request({url: `${baseUrl}/create`, method: 'POST', data})
}

export const apiNormalApiTokenDelete = (data) => {
    return request.request({url: `${baseUrl}/delete`, method: 'POST', data})
}

export const apiNormalApiTokenPage = (data) => {
    return request.request({url: `${baseUrl}/page`, method: 'POST', data})
}
