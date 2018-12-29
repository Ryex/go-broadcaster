import Vue from 'vue'
import Router from 'vue-router'
import Home from '@/components/home'
import { ifNotAuthenticated, ifAuthenticated } from './auth.js'

Vue.use(Router)


export default new Router({
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home,
      beforeEnter: ifAuthenticated,
    },
    {
      path: '/login',
      name: 'Login',
      component: Login,
      beforeEnter: ifNotAuthenticated,
    },
  ]
})
