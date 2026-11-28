import request from "../../request/index.js";

const baseUrl = '/v1/admin/gost/node/group'

export const apiAdminNodeGroupCreate = (data) => {
    return request.request({url: `${baseUrl}/create`, method: 'POST', data})
}

export const apiAdminNodeGroupUpdate = (data) => {
    return request.request({url: `${baseUrl}/update`, method: 'POST', data})
}

export const apiAdminNodeGroupDelete = (data) => {
    return request.request({url: `${baseUrl}/delete`, method: 'POST', data})
}

export const apiAdminNodeGroupPage = (data) => {
    return request.request({url: `${baseUrl}/page`, method: 'POST', data})
}
