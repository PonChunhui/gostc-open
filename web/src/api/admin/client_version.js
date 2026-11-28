import request from "../../request/index.js";

const baseUrl = '/v1/admin/client/version'

export const apiAdminClientVersionCreate = (data) => {
    return request.request({url: `${baseUrl}/create`, method: 'POST', data})
}

export const apiAdminClientVersionUpdate = (data) => {
    return request.request({url: `${baseUrl}/update`, method: 'POST', data})
}

export const apiAdminClientVersionDelete = (data) => {
    return request.request({url: `${baseUrl}/delete`, method: 'POST', data})
}

export const apiAdminClientVersionPage = (data) => {
    return request.request({url: `${baseUrl}/page`, method: 'POST', data})
}
