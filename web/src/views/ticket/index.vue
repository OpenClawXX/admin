<template>
  <div class="ticket-page">
    <a-card :bordered="false">
      <template #title>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>工单列表</span>
          <a-button type="primary" @click="openCreate">
            <template #icon><icon-plus /></template>
            创建工单
          </a-button>
        </div>
      </template>

      <a-space style="margin-bottom: 16px">
        <a-select v-model="query.status" placeholder="状态" allow-clear style="width: 120px">
          <a-option value="created">待派发</a-option>
          <a-option value="assigned">已派发</a-option>
          <a-option value="processing">处理中</a-option>
          <a-option value="suspended">已挂起</a-option>
          <a-option value="review">待验收</a-option>
          <a-option value="completed">已完单</a-option>
        </a-select>
        <a-select v-model="query.priority" placeholder="优先级" allow-clear style="width: 100px">
          <a-option value="p0">紧急</a-option>
          <a-option value="p1">重大</a-option>
          <a-option value="p2">严重</a-option>
          <a-option value="p3">普通</a-option>
        </a-select>
        <a-select v-model="query.type" placeholder="类型" allow-clear style="width: 100px">
          <a-option value="fault">故障</a-option>
          <a-option value="implement">实施</a-option>
          <a-option value="patrol">巡检</a-option>
        </a-select>
        <a-select v-model="query.project_id" placeholder="所属项目" allow-clear style="width: 160px">
          <a-option v-for="p in projectList" :key="p.id" :value="p.id">{{ p.code }} {{ p.name }}</a-option>
        </a-select>
        <a-input-search
          v-model="query.keyword"
          placeholder="搜索工单"
          style="width: 200px"
          @search="fetchTickets"
        />
      </a-space>

      <a-table :columns="columns" :data="tickets" :pagination="pagination" @page-change="onPageChange">
        <template #priority="{ record }">
          <a-tag :color="priorityColor(record.priority)">{{ priorityText(record.priority) }}</a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
        </template>
        <template #type="{ record }">
          {{ typeText(record.type) }}
        </template>
        <template #project="{ record }">
          {{ getProjectName(record.project_id) }}
        </template>
        <template #action="{ record }">
          <a-space>
            <a-link @click="goDetail(record.id)">查看</a-link>
            <a-link v-if="isAdmin" status="danger" @click="handleDeleteTicket(record.id)">删除</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <!-- Create Ticket Modal -->
    <a-modal v-model:visible="showCreateModal" title="创建工单" @ok="handleCreate" :width="600">
      <a-form :model="createForm" layout="vertical">
        <a-form-item field="title" label="标题" :rules="[{ required: true }]">
          <a-input v-model="createForm.title" />
        </a-form-item>
        <a-form-item field="description" label="描述">
          <a-textarea v-model="createForm.description" :max-length="2000" show-word-limit />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item field="type" label="类型" :rules="[{ required: true }]">
              <a-select v-model="createForm.type">
                <a-option value="fault">故障</a-option>
                <a-option value="implement">实施</a-option>
                <a-option value="patrol">巡检</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="priority" label="优先级" :rules="[{ required: true }]">
              <a-select v-model="createForm.priority">
                <a-option value="p0">紧急</a-option>
                <a-option value="p1">重大</a-option>
                <a-option value="p2">严重</a-option>
                <a-option value="p3">普通</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="assignee_id" label="指派给">
              <a-select
                v-model="createForm.assignee_id"
                :placeholder="isEngineer ? '指派给自己' : '选择工程师'"
                :disabled="isEngineer"
                allow-clear
              >
                <a-option v-for="eng in engineers" :key="eng.id" :value="eng.id">
                  {{ eng.real_name || eng.username }}
                </a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="project_id" label="所属项目">
          <a-select v-model="createForm.project_id" placeholder="选择所属项目（可选）" allow-clear>
            <a-option v-for="p in projectList" :key="p.id" :value="p.id">{{ p.code }} - {{ p.name }}</a-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getTicketList, createTicket, deleteTicket } from '@/api/ticket'
import { getUserList } from '@/api/user'
import request from '@/utils/request'
import type { TicketInfo } from '@/api/ticket'
import { useUserStore } from '@/stores/user'
import { Message, Modal } from '@arco-design/web-vue'

const router = useRouter()
const userStore = useUserStore()
const tickets = ref<TicketInfo[]>([])
const engineers = ref<any[]>([])
const projectList = ref<any[]>([])
const showCreateModal = ref(false)

const isEngineer = computed(() => userStore.userInfo?.role === 'engineer')
const isSupervisor = computed(() => userStore.userInfo?.role === 'supervisor')
const isAdmin = computed(() => userStore.userInfo?.role === 'admin')
const currentUserId = computed(() => userStore.userInfo?.id || 0)

const query = reactive({
  status: '',
  priority: '',
  type: '',
  project_id: undefined as number | undefined,
  keyword: '',
  page: 1,
  page_size: 20,
})

const pagination = reactive({
  total: 0,
  current: 1,
  pageSize: 20,
})

const createForm = reactive({
  title: '',
  description: '',
  type: 'fault',
  priority: 'p2',
  assignee_id: undefined as number | undefined,
  project_id: undefined as number | undefined,
})

const columns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '标题', dataIndex: 'title', ellipsis: true },
  { title: '类型', dataIndex: 'type', slotName: 'type', width: 70 },
  { title: '优先级', dataIndex: 'priority', slotName: 'priority', width: 70 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 80 },
  { title: '所属项目', dataIndex: 'project_id', slotName: 'project', width: 130 },
  { title: '创建时间', dataIndex: 'created_at', width: 170 },
  { title: '操作', slotName: 'action', width: 100 },
]

const priorityColor = (p: string) => ({ p0: 'red', p1: 'orange', p2: 'gold', p3: 'blue' }[p] || 'gray')
const priorityText = (p: string) => ({ p0: '紧急', p1: '重大', p2: '严重', p3: '普通' }[p] || p)
const statusColor = (s: string) => ({
  created: 'gray', assigned: 'blue', processing: 'cyan',
  suspended: 'orange', review: 'purple', completed: 'green', archived: 'gray'
}[s] || 'gray')
const statusText = (s: string) => ({
  created: '待派发', assigned: '已派发', processing: '处理中',
  suspended: '已挂起', review: '待验收', completed: '已完单', archived: '已归档'
}[s] || s)
const typeText = (t: string) => ({ fault: '故障', implement: '实施', patrol: '巡检' }[t] || t)

const getProjectName = (id: number | null) => {
  if (!id) return '-'
  const p = projectList.value.find((p: any) => p.id === id)
  return p ? p.name : `ID:${id}`
}

const fetchProjects = async () => {
  try {
    projectList.value = (await request.get('/projects') as any) || []
  } catch (e) {}
}

const fetchTickets = async () => {
  try {
    const result = await getTicketList(query)
    tickets.value = result.list || []
    pagination.total = result.total
    pagination.current = result.page
  } catch (e) {
    // handled by interceptor
  }
}

const onPageChange = (page: number) => {
  query.page = page
  fetchTickets()
}

const goDetail = (id: number) => {
  router.push(`/tickets/${id}`)
}

const handleCreate = async () => {
  try {
    await createTicket(createForm)
    Message.success('工单创建成功')
    showCreateModal.value = false
    createForm.title = ''
    createForm.description = ''
    fetchTickets()
  } catch (e) {
    // handled
  }
}

const handleDeleteTicket = (id: number) => {
  Modal.confirm({
    title: '确认删除',
    content: '删除后不可恢复，确定要删除该工单吗？',
    onOk: async () => {
      await deleteTicket(id)
      Message.success('工单已删除')
      fetchTickets()
    },
  })
}

const fetchEngineers = async () => {
  try {
    const params: any = { page: 1, page_size: 100 }
    if (isSupervisor.value && userStore.userInfo?.team_id) {
      params.team_id = userStore.userInfo.team_id
    }
    const result = await getUserList(params) as any
    engineers.value = result?.list || []
  } catch (e) {}
}

const openCreate = () => {
  createForm.title = ''
  createForm.description = ''
  createForm.type = 'fault'
  createForm.priority = 'p2'
  createForm.project_id = undefined
  if (isEngineer.value) {
    createForm.assignee_id = currentUserId.value
  } else {
    createForm.assignee_id = undefined
  }
  showCreateModal.value = true
}

onMounted(() => {
  fetchTickets()
  fetchEngineers()
  fetchProjects()
})
</script>

<style scoped>
.ticket-page {
  padding: 16px;
}
</style>
