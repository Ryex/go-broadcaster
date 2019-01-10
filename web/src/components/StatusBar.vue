

<script>
  import {AUTH_LOGOUT} from 'actions/auth'
  export default {
    name: 'login',
    data () {
      return {
        username: 'dogo',
        password: 'dogy',
      }
    },
    methods: {
      logout: function () {
        this.$store.dispatch(AUTH_LOGOUT)
        .then(() => {
          this.$router.push('/login')
        })
      }
    },
  }
</script>
