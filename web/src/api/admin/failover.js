import request from "../../request/index.js";

const baseUrl = '/v1/admin/failover'

export const apiAdminFailoverLogPage = (data) => {
    return request.request({url: `${baseUrl}/log/page`, method: 'POST', data})
}
