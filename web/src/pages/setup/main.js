import Vue from 'vue'
import '@/plugins/axios'
import '@/plugins/vuetify'

import SetupApp from '@/pages/setup/App.vue'

import i18n from '@/i18n'

Vue.config.productionTip = false
Vue.config.devtools = true

Vue.config.productionTip = false

new Vue({
  render: h => h(SetupApp),
  i18n,

  created: function () {

  }
}).$mount('#app')
