<template>
  <div class="login-container">
    <a-card class="login-card" :bordered="false">
      <template #title>
        <h2>运维人员管理平台</h2>
      </template>
      <a-form :model="form" @submit-success="handleLogin" layout="vertical">
        <a-form-item field="username" label="用户名" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model="form.username" placeholder="请输入用户名" size="large">
            <template #prefix><icon-user /></template>
          </a-input>
        </a-form-item>
        <a-form-item field="password" label="密码" :rules="[{ required: true, message: '请输入密码' }]">
          <a-input-password v-model="form.password" placeholder="请输入密码" size="large">
            <template #prefix><icon-lock /></template>
          </a-input-password>
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" long size="large" :loading="loading">
            登录
          </a-button>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { Message } from '@arco-design/web-vue'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const handleLogin = async () => {
  loading.value = true
  try {
    await userStore.login(form.username, form.password)
    Message.success('登录成功')
    router.push('/dashboard')
  } catch (error) {
    // error already handled by interceptor
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 420px;
}
.login-card h2 {
  text-align: center;
  margin: 0;
  color: #1d2129;
}
</style>
