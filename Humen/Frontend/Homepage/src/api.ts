import axios from 'axios'

const api = axios.create({
  baseURL: '/api'
})

// 自動帶 token（只要 token 存在 localStorage）
const token = localStorage.getItem('token')
if (token) {
  api.defaults.headers.common['Authorization'] = 'Bearer ' + token
}

export default api
