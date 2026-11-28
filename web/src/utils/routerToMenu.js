// 分组图标（用户菜单的可折叠子菜单）
const groupIcons = {
    '监控中心': '<svg class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="200" height="200"><path d="M128 832h768v64H64V128h64v704z m128-192l192-192 128 128 256-288 48 42-300 338-132-132-146 146-46-42z" fill="#888888"></path></svg>',
    '开发工具': '<svg class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="200" height="200"><path d="M704 128c-106 0-192 86-192 192 0 24 4 47 12 68L128 784v112h160v-80h80v-80h80l100-100c21 8 44 12 68 12 106 0 192-86 192-192s-86-192-192-192z m48 192c-27 0-48-21-48-48s21-48 48-48 48 21 48 48-21 48-48 48z" fill="#888888"></path></svg>',
}

export function routerToMenu(children, funcMap) {
    let result = []
    let groupIndex = new Map()
    children.forEach(item => {
        if (item.meta?.hidden === 1) {
            return
        }
        if (item.meta?.funcKey) {
            if (funcMap.get(item.meta?.funcKey) !== '1') {
                return
            }
        }
        let data = {
            label: item.meta?.title,
            key: item.name,
            iconSvg: item.meta?.icon,
            show: item.meta?.hidden === 2,
            link: item.meta?.link,
        }
        if (item.children) {
            let ch = routerToMenu(item.children, funcMap)
            if (ch.length !== 0) {
                data.children = ch
            }
        }
        // 同组（meta.group）的菜单项收进一个可折叠子菜单，保持首次出现的位置
        const g = item.meta?.group
        if (g) {
            if (!groupIndex.has(g)) {
                const parent = {
                    label: g,
                    key: 'group:' + g,
                    iconSvg: groupIcons[g] || item.meta?.icon,
                    show: true,
                    children: [],
                }
                groupIndex.set(g, parent)
                result.push(parent)
            }
            groupIndex.get(g).children.push(data)
        } else {
            result.push(data)
        }
    })
    return result
}
