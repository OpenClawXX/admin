import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, getProfile } from '@/api/user'
import type { UserInfo } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)

  const login = async (username: string, password: string) => {
    const result = await loginApi({ username, password })
    token.value = result.token
    userInfo.value = result.user
    localStorage.setItem('token', result.token)
    return result
  }

  const fetchProfile = async () => {
    const user = await getProfile()
    userInfo.value = user as unknown as UserInfo
    return user
  }

  const logout = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  return { token, userInfo, login, fetchProfile, logout }
})
