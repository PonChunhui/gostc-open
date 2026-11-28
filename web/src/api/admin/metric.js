import request from "../../request/index.js";

const baseUrl = '/v1/admin/metric'

export const apiAdminMetricRealtime = (data) => {
    return request.request({url: `${baseUrl}/realtime`, method: 'POST', data})
}

export const apiAdminMetricHistory = (data) => {
    return request.request({url: `${baseUrl}/history`, method: 'POST', data})
}

export const apiAdminMetricTop = (data) => {
    return request.request({url: `${baseUrl}/top`, method: 'POST', data})
}
