import request from "../../request/index.js";

const baseUrl = '/v1/admin/system/cdk'

export const apiAdminSystemCdkBatchCreate = (data) => {
    return request.request({
        url: `${baseUrl}/batch/create`,
        method: 'POST',
        data
    })
}

export const apiAdminSystemCdkBatchPage = (data) => {
    return request.request({
        url: `${baseUrl}/batch/page`,
        method: 'POST',
        data
    })
}

export const apiAdminSystemCdkBatchDelete = (data) => {
    return request.request({
        url: `${baseUrl}/batch/delete`,
        method: 'POST',
        data
    })
}

export const apiAdminSystemCdkCardPage = (data) => {
    return request.request({
        url: `${baseUrl}/card/page`,
        method: 'POST',
        data
    })
}

export const apiAdminSystemCdkCardExport = (data) => {
    return request.request({
        url: `${baseUrl}/card/export`,
        method: 'POST',
        data
    })
}

export const apiAdminSystemCdkCardBan = (data) => {
    return request.request({
        url: `${baseUrl}/card/ban`,
        method: 'POST',
        data
    })
}