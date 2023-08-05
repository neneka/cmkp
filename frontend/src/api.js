import axios from 'axios'
import jwtDecode from 'jwt-decode'
import localforage from 'localforage'
import Vue from 'vue'
import VueApollo from 'vue-apollo'
import { createApolloClient } from 'vue-cli-plugin-apollo/graphql-client'
import store from './store'
import { onLogin, onLogout } from './vue-apollo'
import getRole from './gql/getRole.graphql'

Vue.use(VueApollo)
localforage.config({ name: 'cmkp' })

const getToken = () => localStorage.getItem('api_token')
const deleteToken = () => localStorage.removeItem('api_token')
const saveToken = (token) => localStorage.setItem('api_token', token)

let baseUrl
if (process.env.NODE_ENV === 'production') {
  baseUrl = process.env.VUE_APP_API_ENDPOINT + '/api/'
} else {
  baseUrl = process.env.VUE_APP_API_ENDPOINT || 'http://localhost:5000/api/'
}

const { apolloClient, wsClient } = createApolloClient({
  httpEndpoint: baseUrl + 'graphql',
  persisting: false,
  websocketsOnly: false,
  ssr: false,
  getAuth: _ => `Bearer ${getToken()}`
})
apolloClient.wsClient = wsClient

let client
const initClient = async (token) => {
  if (token) {
    let decoded
    try {
      decoded = jwtDecode(token)
    } catch (e) {
      console.error(e)
      deleteToken()
      await onLogout(apolloClient)
      return
    }
    store.commit('setUserId', parseInt(decoded.sub, 10))
    store.commit('setName', decoded.aud)

    client = axios.create({
      baseURL: baseUrl,
      headers: {
        'Authorization': 'Bearer ' + token
      }
    })
    await onLogin(apolloClient)
    const { data } = await apolloClient.query({
      query: getRole,
      fetchPolicy: 'network-only'
    })
    store.commit('setRole', data.me.role)
  } else {
    client = axios.create({
      baseURL: baseUrl
    })
    await onLogout(apolloClient)
  }
}

initClient(getToken())
export default {
  apolloClient,
  apolloProvider: () =>
    new VueApollo({
      defaultClient: apolloClient,
      defaultOptions: {
        $query: {
          fetchPolicy: 'cache-and-network'
        }
      },
      errorHandler (error) {
        // eslint-disable-next-line no-console
        console.error('%cError', 'background: red; color: white; padding: 2px 4px; border-radius: 3px; font-weight: bold;', error.message)
        if (error.message && error.message.includes('401')) {
          deleteToken()
          location.reload()
        }
      }
    }),
  login: async (id, password) => {
    const res = await client.post('/login', {
      username: id,
      password: password
    })
    const token = res.data.token

    saveToken(token)
    await initClient(token)
  },
  logout: async () => {
    deleteToken()
    await initClient(null)
  }
}
