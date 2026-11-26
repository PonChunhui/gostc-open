export function requiredRule(tip = '不能为空') {
    return {
        required: true,
        trigger: ["blur", "input"],
        message: tip
    }
}

// 数字输入框(n-input-number)专用必填规则。
// requiredRule 默认按 string 类型校验，数字值会被判为类型错误，
// 又因 message 覆盖而误显示成"不能为空"，故数字字段必须显式 type:'number'。
export function requiredNumberRule(tip = '不能为空') {
    return {
        required: true,
        type: 'number',
        trigger: ["blur", "change"],
        message: tip
    }
}

export function regexpRule(regexp = (str) => {
    return true
}, tip = '格式错误') {
    return {
        required: true,
        trigger: ["blur", "change"],
        validator: (rule, value) => {
            if (!value) {
                return new Error('请输入')
            }
            if (!regexp(value)) {
                return new Error(tip)
            }
            return true
        }
    }
}