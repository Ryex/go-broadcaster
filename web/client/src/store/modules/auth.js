
/*
  Client side Auth logic and state
*/

import axios from 'axios'

export const auth = {
  state: {
    token:  '',
    status: '',
  },
  getters: {
    isAuthenticated: state => !!state.token,
    authStatus: state => state.status,
    userToken: state => state.token,
  },
  actions: {
    'AUTH_REQUEST': ({commit, dispatch}, user) => {
      return new Promise((resolve, reject) => { // The Promise used for router redirect in login
        commit('AUTH_REQUEST')
        axios({url: 'auth', data: user, method: 'POST' })
          .then(resp => {
            const token = resp.data.token
            axios.defaults.headers.common['Authorization'] = token
            commit('AUTH_SUCCESS', token)
            // you have your token, now log in your user :)
            dispatch('USER_REQUEST')
            resolve(resp)
          })
        .catch(err => {
          commit('AUTH_ERROR', err)
          reject(err)
        })
      })
    },
    'AUTH_LOGOUT': ({commit, dispatch}) => {
      return new Promise((resolve, reject) => {
        commit('AUTH_LOGOUT')
        resolve()
      })
    }
  },
  mutations: {
    'AUTH_REQUEST': (state) => {
      state.status = 'loading'
    },
    'AUTH_SUCCESS': (state, token) => {
      state.status = 'success'
      state.token = token
    },
    'AUTH_ERROR': (state) => {
      state.status = 'error'
      state.token = ''
    },
  }

}
