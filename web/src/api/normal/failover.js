import request from "../../request/index.js";

const baseUrl = '/v1/normal/failover'

export const apiFailoverLogPage = (data) => {
    return request.request({url: `${baseUrl}/log/page`, method: 'POST', data})
}
