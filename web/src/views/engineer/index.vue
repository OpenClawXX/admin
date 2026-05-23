<template>
  <div class="engineer-page">
    <a-card :bordered="false">
      <template #title>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>工程师管理</span>
          <a-button type="primary" @click="openCreate">
            <template #icon><icon-plus /></template>
            添加用户
          </a-button>
        </div>
      </template>

      <a-table :columns="columns" :data="users" :pagination="pagination" @page-change="onPageChange">
        <template #role="{ record }">
          <a-tag :color="roleColor(record.role)">{{ roleText(record.role) }}</a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="record.status === 1 ? 'green' : 'red'">{{ record.status === 1 ? '正常' : '禁用' }}</a-tag>
        </template>
        <template #action="{ record }">
          <a-space>
            <a-link @click="editUser(record)">编辑</a-link>
            <a-link @click="openResetModal(record)">重置密码</a-link>
            <a-link status="danger" @click="handleDelete(record.id)">删除</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- Create/Edit Modal -->
    <a-modal v-model:visible="showCreateModal" :title="editingUser ? '编辑用户' : '添加用户'" @ok="handleSubmit" @cancel="resetForm" :width="500">
      <a-form :model="form" layout="vertical">
        <a-form-item field="username" label="用户名" :rules="[{ required: true }]" :validate-trigger="['blur']">
          <a-input v-model="form.username" :disabled="!!editingUser" />
        </a-form-item>
        <a-form-item v-if="!editingUser" field="password" label="密码" :rules="[{ required: true }]">
          <a-input-password v-model="form.password" />
        </a-form-item>
        <a-form-item field="real_name" label="姓名">
          <a-input v-model="form.real_name" />
        </a-form-item>
        <a-form-item field="email" label="邮箱">
          <a-input v-model="form.email" />
        </a-form-item>
        <a-form-item field="phone" label="手机号">
          <a-input v-model="form.phone" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item field="role" label="角色" :rules="[{ required: true }]">
              <a-select v-model="form.role">
                <a-option value="admin">管理员</a-option>
                <a-option value="supervisor">主管</a-option>
                <a-option value="engineer">工程师</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="team_id" label="团队">
              <a-select v-model="form.team_id" allow-clear placeholder="选择团队">
                <a-option v-for="team in teams" :key="team.id" :value="team.id">{{ team.name }}</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="status" label="状态">
              <a-switch :model-value="form.status === 1" @change="(val: boolean) => form.status = val ? 1 : 0">
                <template #checked>正常</template>
                <template #unchecked>禁用</template>
              </a-switch>
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>

    <!-- Reset Password Modal -->
    <a-modal
      v-model:visible="showResetModal"
      title="重置密码"
      :ok-text="resetResult ? '关闭' : '确认重置'"
      :cancel="() => { showResetModal = false }"
      @ok="resetResult ? showResetModal = false : handleResetPassword()"
    >
      <a-form layout="vertical">
        <a-form-item label="用户">
          <a-input :model-value="resettingUser?.real_name || resettingUser?.username" disabled />
        </a-form-item>
        <a-form-item label="新密码">
          <a-input-password
            v-model="resetPasswordInput"
            placeholder="留空则自动生成8位强密码"
          />
        </a-form-item>
      </a-form>
      <a-alert v-if="resetResult" type="success" style="margin-top: 12px">
        <template #title>密码重置成功</template>
        新密码为：<a-typography-text copyable strong>{{ resetResult }}</a-typography-text>
      </a-alert>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { getUserList, createUser, updateUser, deleteUser, resetPassword, getTeamList } from '@/api/user'
import { Message, Modal } from '@arco-design/web-vue'

const users = ref<any[]>([])
const teams = ref<any[]>([])
const showCreateModal = ref(false)
const showResetModal = ref(false)
const editingUser = ref<any>(null)
const resettingUser = ref<any>(null)
const resetPasswordInput = ref('')
const resetResult = ref('')

const pagination = reactive({ total: 0, current: 1, pageSize: 20 })

const form = reactive({
  username: '',
  password: '',
  real_name: '',
  email: '',
  phone: '',
  role: 'engineer',
  team_id: undefined as number | undefined,
  status: 1,
})

const columns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '用户名', dataIndex: 'username', width: 120 },
  { title: '姓名', dataIndex: 'real_name', width: 100 },
  { title: '邮箱', dataIndex: 'email', width: 180 },
  { title: '角色', dataIndex: 'role', slotName: 'role', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 80 },
  { title: '操作', slotName: 'action', width: 180 },
]

const roleColor = (r: string) => ({ admin: 'red', supervisor: 'orange', engineer: 'blue' }[r] || 'gray')
const roleText = (r: string) => ({ admin: '管理员', supervisor: '主管', engineer: '工程师' }[r] || r)

const fetchUsers = async () => {
  const result = await getUserList({ page: pagination.current, page_size: pagination.pageSize }) as any
  users.value = result?.list || []
  pagination.total = result?.total || 0
}

const fetchTeams = async () => {
  teams.value = (await getTeamList() as any) || []
}

const onPageChange = (page: number) => {
  pagination.current = page
  fetchUsers()
}

const openCreate = () => {
  resetForm()
  showCreateModal.value = true
}

const editUser = (user: any) => {
  editingUser.value = user
  Object.assign(form, {
    username: user.username,
    password: '',
    real_name: user.real_name,
    email: user.email,
    phone: user.phone,
    role: user.role,
    team_id: user.team_id,
    status: user.status,
  })
  showCreateModal.value = true
}

const resetForm = () => {
  editingUser.value = null
  Object.assign(form, { username: '', password: '', real_name: '', email: '', phone: '', role: 'engineer', team_id: undefined, status: 1 })
}

const handleSubmit = async () => {
  try {
    if (editingUser.value) {
      await updateUser(editingUser.value.id, form)
      Message.success('更新成功')
    } else {
      await createUser(form)
      Message.success('创建成功')
    }
    showCreateModal.value = false
    resetForm()
    fetchUsers()
  } catch (e) {
    // handled
  }
}

const handleDelete = (id: number) => {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该用户吗？',
    onOk: async () => {
      await deleteUser(id)
      Message.success('删除成功')
      fetchUsers()
    },
  })
}

const openResetModal = (user: any) => {
  resettingUser.value = user
  resetPasswordInput.value = ''
  resetResult.value = ''
  showResetModal.value = true
}

const handleResetPassword = async () => {
  if (!resettingUser.value) return
  try {
    const result = await resetPassword(resettingUser.value.id, resetPasswordInput.value)
    resetResult.value = result.password
    Message.success('密码重置成功')
  } catch (e) {}
}

onMounted(() => {
  fetchUsers()
  fetchTeams()
})
</script>

<style scoped>
.engineer-page {
  padding: 16px;
}
</style>
