import request from "../../request/index.js";

const baseUrl = '/v1/normal/metric'

export const apiNormalMetricRealtime = (data) => {
    return request.request({url: `${baseUrl}/realtime`, method: 'POST', data})
}

export const apiNormalMetricHistory = (data) => {
    return request.request({url: `${baseUrl}/history`, method: 'POST', data})
}
