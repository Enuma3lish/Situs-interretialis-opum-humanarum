<template>
  <main>
    <div v-if="!token" style="max-width:400px;margin:auto">
      <h2>請先登入</h2>
      <div>
        <label>帳號：</label>
        <input v-model="loginUser" />
      </div>
      <div>
        <label>密碼：</label>
        <input v-model="loginPass" type="password" />
      </div>
      <button @click="login">登入</button>
      <button @click="register">註冊</button>
      <div v-if="loginError" style="color:red;">{{ loginError }}</div>
      <div style="font-size:14px;color:#888;margin-top:10px;">
        預設管理員：帳號 admin，密碼 root
      </div>
    </div>

    <div v-else>
      <h1>職缺列表</h1>
      <div>
        <input v-model="search" placeholder="請輸入職缺關鍵字" />
        <button @click="fetchJobs">搜尋</button>
        <button @click="fetchAvg">顯示公司平均/高薪</button>
        <button @click="logout" style="float:right">登出</button>
      </div>
      <br />
      <table v-if="jobs.length > 0" border="1" cellpadding="6">
        <thead>
          <tr>
            <th>公司</th>
            <th>職缺名稱</th>
            <th>薪資下限</th>
            <th>薪資上限</th>
            <th v-if="isAdmin">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="job in jobs" :key="job.id">
            <td>{{ job.company }}</td>
            <td>{{ job.title }}</td>
            <td>{{ job.salary_min }}</td>
            <td>{{ job.salary_max }}</td>
            <td v-if="isAdmin">
              <button @click="deleteJob(job.id)">刪除</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else>
        <p>尚無職缺</p>
      </div>
      <div v-if="stats.length" style="margin-top:2em">
        <h2>公司統計資訊</h2>
        <ul>
          <li v-for="stat in stats" :key="stat.company">
            {{ stat.company }} 平均薪資 ${{ stat.avg_salary }}，高薪職缺 {{ stat.high_salary }} 筆
          </li>
        </ul>
      </div>
      <div v-if="isAdmin" style="margin-top:3em;">
        <h2>帳號管理（Admin 專屬）</h2>
        <button @click="fetchUsers">載入帳號清單</button>
        <table v-if="users.length" border="1" cellpadding="6" style="margin-top:10px;">
          <thead>
            <tr>
              <th>ID</th>
              <th>帳號</th>
              <th>管理員</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td>{{ user.id }}</td>
              <td>{{ user.username }}</td>
              <td>{{ user.is_admin ? "是" : "" }}</td>
              <td>
                <button @click="deleteUser(user.id)" :disabled="user.id === myUserId">
                  刪除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </main>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import api from './api'

interface Job {
  id: number
  company: string
  title: string
  salary_min: number
  salary_max: number
}

interface Stat {
  company: string
  avg_salary: number
  high_salary: number
}

interface User {
  id: number
  username: string
  is_admin: boolean
}

const jobs = ref<Job[]>([])
const stats = ref<Stat[]>([])
const users = ref<User[]>([])
const token = ref(localStorage.getItem('token') || '')
const loginUser = ref('admin')
const loginPass = ref('root')
const loginError = ref('')
const search = ref('')

const isAdmin = ref(false)
const myUserId = ref<number|null>(null)

const setToken = (t: string, is_admin = false, user_id: number|null = null) => {
  token.value = t
  localStorage.setItem('token', t)
  api.defaults.headers.common['Authorization'] = 'Bearer ' + t
  isAdmin.value = is_admin
  myUserId.value = user_id
}

if (token.value) {
  api.defaults.headers.common['Authorization'] = 'Bearer ' + token.value
  // 取得自己的 isAdmin, userId（用 jwt decode，或每次登入回傳 is_admin/user_id，這裡簡化直接用 api）
  fetchMyInfo()
  fetchJobs()
}

async function fetchMyInfo() {
  // 這個端點需要你在 Go backend 實作，或可在 login 時記下 user_id/is_admin
  // 簡單做法：登入時回傳 token、is_admin、user_id
  // 這邊先省略
}

async function login() {
  loginError.value = ''
  try {
    const res = await api.post('/login', {
      username: loginUser.value,
      password: loginPass.value
    })
    setToken(res.data.token, res.data.is_admin, res.data.user_id)
    fetchJobs()
  } catch (err: any) {
    loginError.value = err?.response?.data?.error || '登入失敗'
  }
}

async function register() {
  loginError.value = ''
  try {
    await api.post('/register', {
      username: loginUser.value,
      password: loginPass.value
    })
    login()
  } catch (err: any) {
    loginError.value = err?.response?.data?.error || '註冊失敗'
  }
}

function logout() {
  token.value = ''
  localStorage.removeItem('token')
  delete api.defaults.headers.common['Authorization']
  jobs.value = []
  stats.value = []
  users.value = []
  isAdmin.value = false
  myUserId.value = null
}

async function fetchJobs() {
  stats.value = []
  let url = '/jobs'
  if (search.value.trim()) {
    url += '?keyword=' + encodeURIComponent(search.value.trim())
  }
  try {
    const res = await api.get<Job[]>(url)
    jobs.value = res.data
  } catch (err) {
    jobs.value = []
    alert('載入職缺失敗！')
  }
}

async function fetchAvg() {
  try {
    const res = await api.get<Stat[]>('/companies/stat')
    stats.value = res.data
  } catch (err) {
    stats.value = []
    alert('載入公司統計失敗！')
  }
}

async function deleteJob(id: number) {
  if (!confirm("確定要刪除此職缺？")) return;
  try {
    await api.delete(`/jobs/${id}`)
    fetchJobs()
  } catch (err) {
    alert("刪除職缺失敗")
  }
}

async function fetchUsers() {
  try {
    const res = await api.get<User[]>('/users')
    users.value = res.data
  } catch (err) {
    users.value = []
    alert("載入帳號清單失敗")
  }
}

async function deleteUser(id: number) {
  if (!confirm("確定要刪除此帳號？")) return;
  try {
    await api.delete(`/users/${id}`)
    users.value = users.value.filter(u => u.id !== id)
  } catch (err) {
    alert("刪除帳號失敗")
  }
}
</script>

<style>
body { font-family: sans-serif; }
main { max-width: 700px; margin: 32px auto; }
table { width: 100%; }
th, td { text-align: center; }
input[type=password] { font-family: inherit; }
button { margin: 0 2px; }
</style>
