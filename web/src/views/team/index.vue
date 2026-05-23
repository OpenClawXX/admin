<template>
  <div class="team-page">
    <a-card :bordered="false">
      <template #title>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>团队管理</span>
          <a-button type="primary" @click="openCreate">
            <template #icon><icon-plus /></template>
            新建团队
          </a-button>
        </div>
      </template>

      <a-table :columns="columns" :data="teams">
        <template #supervisor="{ record }">
          {{ getUserName(record.supervisor_id) }}
        </template>
        <template #memberCount="{ record }">
          {{ getMemberCount(record.id) }}
        </template>
        <template #action="{ record }">
          <a-space>
            <a-link @click="openEdit(record)">编辑</a-link>
            <a-link status="danger" @click="handleDelete(record.id)">删除</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="showModal"
      :title="editing ? '编辑团队' : '新建团队'"
      @ok="handleSubmit"
      @cancel="resetForm"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item field="name" label="团队名称" :rules="[{ required: true, message: '请输入团队名称' }]">
          <a-input v-model="form.name" placeholder="请输入团队名称" />
        </a-form-item>
        <a-form-item field="supervisor_id" label="团队主管">
          <a-select v-model="form.supervisor_id" placeholder="选择团队主管" allow-clear>
            <a-option v-for="user in supervisors" :key="user.id" :value="user.id">
              {{ user.real_name || user.username }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="description" label="团队描述">
          <a-textarea v-model="form.description" placeholder="请输入团队描述" :max-length="500" show-word-limit />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { getTeamList, createTeam, updateTeam, deleteTeam } from '@/api/user'
import { getUserList } from '@/api/user'
import { Message, Modal } from '@arco-design/web-vue'

const teams = ref<any[]>([])
const users = ref<any[]>([])
const showModal = ref(false)
const editing = ref<any>(null)

const form = reactive({
  name: '',
  supervisor_id: undefined as number | undefined,
  description: '',
})

const supervisors = computed(() =>
  users.value.filter(u => u.role === 'admin' || u.role === 'supervisor')
)

const columns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '团队名称', dataIndex: 'name' },
  { title: '主管', dataIndex: 'supervisor_id', slotName: 'supervisor', width: 120 },
  { title: '成员数', slotName: 'memberCount', width: 80 },
  { title: '描述', dataIndex: 'description', ellipsis: true },
  { title: '操作', slotName: 'action', width: 120 },
]

const getUserName = (id: number | null) => {
  if (!id) return '-'
  const user = users.value.find(u => u.id === id)
  return user ? (user.real_name || user.username) : `ID:${id}`
}

const getMemberCount = (teamId: number) => {
  return users.value.filter(u => u.team_id === teamId).length
}

const fetchTeams = async () => {
  teams.value = (await getTeamList() as any) || []
}

const fetchUsers = async () => {
  const result = await getUserList({ page: 1, page_size: 200 }) as any
  users.value = result?.list || []
}

const openCreate = () => {
  editing.value = null
  resetForm()
  showModal.value = true
}

const openEdit = (team: any) => {
  editing.value = team
  Object.assign(form, {
    name: team.name,
    supervisor_id: team.supervisor_id || undefined,
    description: team.description || '',
  })
  showModal.value = true
}

const resetForm = () => {
  Object.assign(form, { name: '', supervisor_id: undefined, description: '' })
}

const handleSubmit = async () => {
  if (!form.name) {
    Message.warning('请输入团队名称')
    return
  }
  try {
    if (editing.value) {
      await updateTeam(editing.value.id, form)
      Message.success('更新成功')
    } else {
      await createTeam(form)
      Message.success('创建成功')
    }
    showModal.value = false
    resetForm()
    fetchTeams()
  } catch (e) {}
}

const handleDelete = (id: number) => {
  const memberCount = users.value.filter(u => u.team_id === id).length
  const content = memberCount > 0
    ? `该团队下有 ${memberCount} 名成员，删除后成员将变为无团队状态，确认删除？`
    : '确定要删除该团队吗？'

  Modal.confirm({
    title: '确认删除',
    content,
    onOk: async () => {
      await deleteTeam(id)
      Message.success('删除成功')
      fetchTeams()
    },
  })
}

onMounted(() => {
  fetchTeams()
  fetchUsers()
})
</script>

<style scoped>
.team-page {
  padding: 16px;
}
</style>
