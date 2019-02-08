import Vue from 'vue'
import './plugins/vuetify'

import App from './App.vue'

import axios from 'axios'
import i18n from './i18n'

Vue.config.productionTip = false
Vue.config.devtools = true

Vue.config.productionTip = false

new Vue({
  render: h => h(App),
  i18n,

  created: function () {

  }
}).$mount('#app')
