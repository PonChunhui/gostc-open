import request from "../../request/index.js";

const baseUrl = '/v1/normal/template'

export const apiNormalTemplateCreate = (data) => {
    return request.request({url: `${baseUrl}/create`, method: 'POST', data})
}

export const apiNormalTemplateDelete = (data) => {
    return request.request({url: `${baseUrl}/delete`, method: 'POST', data})
}

export const apiNormalTemplatePage = (data) => {
    return request.request({url: `${baseUrl}/page`, method: 'POST', data})
}

export const apiNormalTemplateApply = (data) => {
    return request.request({url: `${baseUrl}/apply`, method: 'POST', data})
}
