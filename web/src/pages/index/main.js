import Vue from 'vue'
import './plugins/axios'
import './plugins/vuetify'
import Vuex from 'vuex'

Vue.use(Vuex)

import App from './App.vue'

import router from './router'
import store from './store'

import axios from 'axios'
import i18n from './i18n'

Vue.config.productionTip = false
Vue.config.devtools = true

Vue.config.productionTip = false

new Vue({
  store,
  router,
  render: h => h(App),
  i18n,

  created: function () {
    axios.interceptors.response.use(undefined, function (err) {
      return new Promise(function (resolve, reject) {
        if (err.status === 401 && err.config && !err.config.__isRetryRequest) {
        // if you ever get an unauthorized, logout the user
          this.$store.dispatch('AUTH_LOGOUT')
        // you can also redirect to /login if needed !
        }
        throw err;
      });
    });
    var token = this.$store.getters.userToken
    if (token) {
      axios.defaults.headers.common['Authorization'] = token
    }
  }
}).$mount('#app')
