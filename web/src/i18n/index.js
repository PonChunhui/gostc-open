import {createI18n} from 'vue-i18n'
import zh from './zh.js'
import en from './en.js'

const LOCALE_KEY = 'gostc_locale'

function detectLocale() {
    try {
        const saved = localStorage.getItem(LOCALE_KEY)
        if (saved === 'zh' || saved === 'en') {
            return saved
        }
    } catch (e) {
        // ignore
    }
    return 'zh'
}

const i18n = createI18n({
    legacy: false,
    globalInjection: true,
    locale: detectLocale(),
    fallbackLocale: 'zh',
    messages: {zh, en},
})

export function setLocale(locale) {
    if (locale !== 'zh' && locale !== 'en') {
        return
    }
    i18n.global.locale.value = locale
    try {
        localStorage.setItem(LOCALE_KEY, locale)
    } catch (e) {
        // ignore
    }
}

export function getLocale() {
    return i18n.global.locale.value
}

export default i18n
