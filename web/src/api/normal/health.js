import request from "../../request/index.js";

const baseUrl = '/v1/normal/gost/client/health'

export const apiNormalHealthPage = (data) => {
    return request.request({url: `${baseUrl}/page`, method: 'POST', data})
}

export const apiNormalHealthOverview = (data) => {
    return request.request({url: `${baseUrl}/overview`, method: 'POST', data})
}
