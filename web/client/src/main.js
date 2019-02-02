import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

import App from './App.vue'

import Router from './router'

Vue.config.productionTip = false
Vue.config.devtools = true

Vue.config.productionTip = false

new Vue({
  render: h => h(App),
  created: function () {
    axios.interceptors.response.use(undefined, function (err) {
      return new Promise(function (resolve, reject) {
        if (err.status === 401 && err.config && !err.config.__isRetryRequest) {
        // if you ever get an unauthorized, logout the user
          this.$store.dispatch(AUTH_LOGOUT)
        // you can also redirect to /login if needed !
        }
        throw err;
      });
    });
    token = this.$store.getters.userToken
    if (token) {
      axios.defaults.headers.common['Authorization'] = token
    }
  }
}).$mount('#app')
