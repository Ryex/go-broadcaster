<template>
 <div>
   <form class="login" @submit.prevent="login">
     <h1>Sign in</h1>
     <label>User name</label>
     <input required v-model="username" type="text" placeholder="Snoopy"/>
     <label>Password</label>
     <input required v-model="password" type="password" placeholder="Password"/>
     <hr/>
     <button type="submit">Login</button>
   </form>
 </div>
</template>

<script>
  export default {
    name: 'Login',
    data () {
      return {
        username: 'dogo',
        password: 'dogy',
      }
    },
    methods: {
      login: function () {
        const { username, password } = this
        this.$store.dispatch('AUTH_REQUEST', { username, password }).then(() => {
          this.$router.push('/')
        })
      }
    },
  }
</script>
