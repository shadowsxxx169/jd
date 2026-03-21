import { createRouter, createWebHistory } from 'vue-router'
import apiBus from '@/utils/apiBus'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // 象棋模块
    {
      path: '/chess',
      component: () => import('@/views/chess/LobbyView.vue'),
    },
    {
      path: '/chess/game',
      component: () => import('@/views/chess/GameView.vue'),
    },
    {
      path: '/',
      name: 'default',
      component: () => import('@/layouts/DefaultLayout.vue'),
      children: [
        {
          path: '', // 添加 /post 路由规则
          component: () => import('@/views/default/PostView.vue'),
        },
        {
          path: 'create',
          component: () => import('@/views/default/CreateView.vue'), // 修改为新创建的组件
          meta: { requiresAuth: true },
        },
      ],
    },
    {
      path: '/auth/',
      name: 'auth',
      component: () => import('@/layouts/AuthLayout.vue'),
      children: [
        {
          path: 'login',
          name: 'login',
          component: () => import('@/views/auth/LoginView.vue'),
        },
        {
          path: 'register',
          name: 'register',
          component: () => import('@/views/auth/RegisterView.vue'),
        },
      ],
    },
  ],
})

// 路由守卫，验证用户是否登录
router.beforeEach((_to, _from, next) => {
  // if (to.meta.requiresAuth && !localStorage.getItem('token')) {
  //   next('/login')
  // } else {
  //   next()
  // }
  next()
})

const logout = () => {
  router.push('/auth/login')
}

apiBus.on('API:LOGIN', (req) => {
  const { stop } = req
  if (stop) return
  router.push('/')
})

apiBus.on('API:UN_AUTH', () => {
  logout()
})

apiBus.on('API:LOGOUT', () => {
  logout()
})

export default router
