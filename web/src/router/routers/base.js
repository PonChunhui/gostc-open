export const baseRouters = [
    {
        path: '/',
        name: 'Home',
        component: () => import('../../views/home/index.vue'),
        meta: {
            title: '首页',
            layout: 'empty',
            hidden: 1,
            icon: ''
        }
    },
    {
        path: '/download',
        name: 'Download',
        component: () => import('../../views/download/index.vue'),
        meta: {
            title: '客户端下载',
            layout: 'empty',
            hidden: 1,
            icon: '',
        }
    },
    {
        path: '/login',
        name: 'Login',
        component: () => import('../../views/login/index.vue'),
        meta: {
            title: '登录',
            layout: 'empty',
            hidden: 1,
            icon: '',
        }
    },
    {
        path: '/resetPassword',
        name: 'ResetPassword',
        component: () => import('../../views/resetPassword/index.vue'),
        meta: {
            title: '忘记密码',
            layout: 'empty',
            hidden: 1,
            icon: '',
        }
    },
    {
        path: '/403',
        name: '403',
        component: () => import('../../views/public/403.vue'),
        meta: {
            title: '403',
            layout: 'empty',
            hidden: 1,
            icon: '',
        }
    },
    {
        path: '/:pathMatch(.*)',
        name: '404',
        component: () => import('../../views/public/404.vue'),
        meta: {
            title: '404',
            layout: 'empty'
        }
    }
]