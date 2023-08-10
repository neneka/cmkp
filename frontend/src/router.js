import Vue from 'vue'
import Router from 'vue-router'
import store from './store'
import Root from './Root'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      component: Root,
      beforeEnter: async function (to, from, next) {
        if (store.getters.loggedIn) {
          next()
        } else {
          next('/login')
        }
      },
      children: [
        {
          path: '',
          name: 'ホーム',
          component: () => import('./views/Home.vue')
        },
        {
          path: 'changePassword',
          name: 'パスワード変更',
          component: () => import('./views/ChangeMyPassword.vue')
        },
        {
          path: 'circles',
          name: 'サークル検索',
          component: () => import('./views/SearchCircle')
        },
        {
          path: 'circles/:cid(\\d+)',
          name: 'サークル詳細',
          component: () => import('./views/CircleInfo'),
          meta: {
            backButton: true
          },
          props: route => ({ cid: parseInt(route.params.cid, 10) })
        },
        {
          path: 'my-requests',
          component: Root,
          children: [
            {
              path: '',
              name: 'マイリクエスト',
              component: () => import('./views/MyRequests')
            },
            {
              path: 'create/:cid(\\d+)',
              name: 'リクエスト作成',
              component: () => import('./views/CreateRequest'),
              meta: {
                backButton: true
              },
              props: route => ({ cid: parseInt(route.params.cid, 10) })
            },
            {
              path: 'notes',
              name: 'リクエスト備考',
              component: () => import('./views/MyRequestNote')
            }
          ]
        },
        {
          path: 'planning',
          component: Root,
          children: [
            {
              path: 'all-requests/:day(\\d+)',
              name: '全リクエスト一覧',
              component: () => import('./views/planning/AllRequestList'),
              props: route => ({ day: parseInt(route.params.day, 10) })
            },
            {
              path: 'users',
              name: 'メンバー別詳細',
              component: () => import('./views/planning/Users')
            },
            {
              path: 'users/:id(\\d+)',
              name: 'メンバー詳細',
              component: () => import('./views/planning/UserDetail'),
              meta: {
                backButton: true
              },
              props: route => ({ userId: parseInt(route.params.id, 10) })
            },
            {
              path: 'all-request-notes',
              name: '全リクエスト備考',
              component: () => import('./views/planning/AllRequestNote')
            },
            {
              path: 'all-request-text',
              name: '全リクエストテキスト',
              component: () => import('./views/planning/AllRequestText')
            }
          ],
          beforeEnter: async function (to, from, next) {
            next()
            // if (store.getters.isPlanner) {
            //   next()
            // } else {
            //   next('/')
            // }
          }
        },
        {
          path: 'admin',
          component: Root,
          children: [
            {
              path: 'users',
              name: 'メンバーリスト',
              component: () => import('./views/admin/UserList')
            },
            {
              path: 'users/create',
              name: 'メンバー登録',
              component: () => import('./views/admin/UserCreate'),
              meta: {
                backPage: '/admin/users'
              }
            },
            {
              path: 'config',
              name: '設定',
              component: () => import('./views/admin/Config')
            }
          ],
          beforeEnter: async function (to, from, next) {
            if (store.getters.isAdmin) {
              next()
            } else {
              next('/')
            }
          }
        }
      ]
    },
    {
      path: '/login',
      name: 'ログイン',
      component: () => import('./views/Login'),
      beforeEnter: async function (to, from, next) {
        if (store.getters.loggedIn) {
          next('/')
        } else {
          next()
        }
      }
    }
  ]
})
