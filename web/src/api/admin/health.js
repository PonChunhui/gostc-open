import request from "../../request/index.js";

const baseUrl = '/v1/admin/gost/client/health'

export const apiAdminHealthPage = (data) => {
    return request.request({url: `${baseUrl}/page`, method: 'POST', data})
}

export const apiAdminHealthOverview = (data) => {
    return request.request({url: `${baseUrl}/overview`, method: 'POST', data})
}
