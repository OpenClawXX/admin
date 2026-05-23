<template>
  <a-layout class="layout">
    <a-layout-sider :collapsed="collapsed" collapsible @collapse="onCollapse" :width="220">
      <div class="logo">
        <h1 v-if="!collapsed">运维管理平台</h1>
        <h1 v-else>运维</h1>
      </div>
      <a-menu
        :selected-keys="selectedKeys"
        :style="{ width: '100%' }"
        @menu-item-click="onMenuClick"
      >
        <a-menu-item key="/dashboard">
          <template #icon><icon-dashboard /></template>
          工作台
        </a-menu-item>
        <a-menu-item key="/tickets">
          <template #icon><icon-file /></template>
          工单管理
        </a-menu-item>
        <a-menu-item v-if="!isEngineer" key="/projects">
          <template #icon><icon-folder /></template>
          项目管理
        </a-menu-item>
        <a-menu-item v-if="isAdminOrSupervisor" key="/engineers">
          <template #icon><icon-user /></template>
          工程师管理
        </a-menu-item>
        <a-menu-item v-if="isAdminOrSupervisor" key="/teams">
          <template #icon><icon-user-group /></template>
          团队管理
        </a-menu-item>
        <a-menu-item key="/knowledge">
          <template #icon><icon-book /></template>
          知识库
        </a-menu-item>
        <a-menu-item v-if="isAdminOrSupervisor" key="/schedule">
          <template #icon><icon-calendar /></template>
          排班管理
        </a-menu-item>
        <a-menu-item v-if="isAdminOrSupervisor" key="/assets">
          <template #icon><icon-desktop /></template>
          资产管理
        </a-menu-item>
        <a-menu-item v-if="isAdmin" key="/system">
          <template #icon><icon-settings /></template>
          系统设置
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header>
        <div class="header-right">
          <a-dropdown>
            <a-button type="text">
              <icon-user /> {{ userStore.userInfo?.real_name || userStore.userInfo?.username }}
              <icon-down />
            </a-button>
            <template #content>
              <a-doption @click="handleLogout">退出登录</a-doption>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>
      <a-layout-content>
        <router-view v-if="profileLoaded" />
        <div v-else style="display:flex;justify-content:center;padding-top:200px">
          <a-spin />
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const collapsed = ref(false)
const profileLoaded = ref(false)

const selectedKeys = computed(() => [route.path])

const role = computed(() => userStore.userInfo?.role || '')
const isAdmin = computed(() => role.value === 'admin')
const isAdminOrSupervisor = computed(() => role.value === 'admin' || role.value === 'supervisor')
const isEngineer = computed(() => role.value === 'engineer')

const onCollapse = (val: boolean) => {
  collapsed.value = val
}

const onMenuClick = (key: string) => {
  router.push(key)
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

onMounted(async () => {
  if (userStore.token) {
    try {
      await userStore.fetchProfile()
    } catch (e) {
      // profile fetch failed, still allow rendering
    }
  }
  profileLoaded.value = true
})
</script>

<style scoped>
.layout {
  height: 100vh;
}
.logo {
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}
.logo h1 {
  font-size: 16px;
  margin: 0;
}
.header-right {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding: 0 24px;
  height: 48px;
  background: #fff;
  border-bottom: 1px solid #e5e6eb;
}
</style>
