<template>
  <div class="project-page">
    <a-card :bordered="false">
      <template #title>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>项目管理</span>
          <a-button v-if="isAdminOrSupervisor" type="primary" @click="openCreate">
            <template #icon><icon-plus /></template>
            创建项目
          </a-button>
        </div>
      </template>

      <a-table :columns="columns" :data="projects">
        <template #status="{ record }">
          <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
        </template>
        <template #type="{ record }">
          {{ typeText(record.type) }}
        </template>
        <template #priority="{ record }">
          <a-tag :color="priorityColor(record.priority)">{{ priorityText(record.priority) }}</a-tag>
        </template>
        <template #manager="{ record }">
          {{ getManagerName(record.manager_id) }}
        </template>
        <template #action="{ record }">
          <a-space>
            <a-link @click="openDetail(record)">详情</a-link>
            <a-link v-if="isAdminOrSupervisor" @click="editProject(record)">编辑</a-link>
            <a-link v-if="isAdminOrSupervisor" status="danger" @click="handleDelete(record.id)">删除</a-link>
          </a-space>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:visible="showModal" :title="editing ? '编辑项目' : '创建项目'" @ok="handleSubmit" @cancel="resetForm" :width="640">
      <a-form :model="form" layout="vertical">
        <a-form-item v-if="editing" label="项目编号">
          <a-input :model-value="editing?.code" disabled />
        </a-form-item>
        <a-form-item field="name" label="项目名称" :rules="[{ required: true, message: '请输入项目名称' }]">
          <a-input v-model="form.name" placeholder="请输入项目名称" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item field="type" label="项目类型">
              <a-select v-model="form.type">
                <a-option value="daily">日常运维</a-option>
                <a-option value="special">专项任务</a-option>
                <a-option value="emergency">应急响应</a-option>
                <a-option value="patrol">巡检项目</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="priority" label="优先级">
              <a-select v-model="form.priority">
                <a-option value="high">高</a-option>
                <a-option value="medium">中</a-option>
                <a-option value="low">低</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="budget" label="预算（元）">
              <a-input-number v-model="form.budget" :min="0" :precision="2" placeholder="可留空" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="requester" label="需求方">
          <a-input v-model="form.requester" placeholder="提出需求的部门或客户" />
        </a-form-item>
        <a-form-item field="manager_id" label="负责人" :rules="[{ required: true, message: '请选择负责人' }]">
          <a-select v-model="form.manager_id" placeholder="选择负责人" filterable>
            <a-option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.real_name || user.username }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="member_ids" label="项目成员">
          <a-select v-model="form.member_ids" placeholder="选择项目成员" multiple filterable>
            <a-option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.real_name || user.username }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item field="start_date" label="开始日期">
              <a-date-picker v-model="form.start_date" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="end_date" label="计划结束日期">
              <a-date-picker v-model="form.end_date" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="actual_end_date" label="实际完成日期">
              <a-date-picker v-model="form.actual_end_date" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="description" label="项目描述">
          <a-textarea v-model="form.description" placeholder="需求描述" :max-length="2000" show-word-limit />
        </a-form-item>
        <a-form-item field="remark" label="备注">
          <a-textarea v-model="form.remark" placeholder="补充说明" :max-length="1000" show-word-limit />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Project Detail Drawer -->
    <a-drawer :visible="showDrawer" @cancel="showDrawer = false" :width="720" :footer="false">
      <template #title>{{ detailProject?.code }} - {{ detailProject?.name }}</template>
      <a-descriptions :column="2" bordered size="small" v-if="detailProject">
        <a-descriptions-item label="项目编号">{{ detailProject.code }}</a-descriptions-item>
        <a-descriptions-item label="项目名称">{{ detailProject.name }}</a-descriptions-item>
        <a-descriptions-item label="类型">{{ typeText(detailProject.type) }}</a-descriptions-item>
        <a-descriptions-item label="优先级">
          <a-tag :color="priorityColor(detailProject.priority)">{{ priorityText(detailProject.priority) }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag :color="statusColor(detailProject.status)">{{ statusText(detailProject.status) }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="负责人">{{ getManagerName(detailProject.manager_id) }}</a-descriptions-item>
        <a-descriptions-item label="需求方">{{ detailProject.requester || '-' }}</a-descriptions-item>
        <a-descriptions-item label="预算">{{ detailProject.budget ? `¥${detailProject.budget}` : '-' }}</a-descriptions-item>
        <a-descriptions-item label="项目成员" :span="2">
          <a-space v-if="detailMembers.length > 0" wrap>
            <a-tag v-for="m in detailMembers" :key="m.id" color="arcoblue">{{ m.real_name || m.username }}</a-tag>
          </a-space>
          <span v-else>-</span>
        </a-descriptions-item>
        <a-descriptions-item label="开始日期">{{ detailProject.start_date || '-' }}</a-descriptions-item>
        <a-descriptions-item label="计划结束">{{ detailProject.end_date || '-' }}</a-descriptions-item>
        <a-descriptions-item label="实际完成">{{ detailProject.actual_end_date || '-' }}</a-descriptions-item>
        <a-descriptions-item label="描述" :span="2">{{ detailProject.description || '-' }}</a-descriptions-item>
        <a-descriptions-item label="备注" :span="2">{{ detailProject.remark || '-' }}</a-descriptions-item>
      </a-descriptions>

      <a-divider>关联工单（{{ projectTickets.length }}）</a-divider>
      <a-table :columns="ticketColumns" :data="projectTickets" :pagination="false" size="small" :scroll="{ y: 300 }">
        <template #ticketType="{ record }">
          {{ ticketTypeText(record.type) }}
        </template>
        <template #priority="{ record }">
          <a-tag :color="ticketPriorityColor(record.priority)" size="small">{{ ticketPriorityText(record.priority) }}</a-tag>
        </template>
        <template #status="{ record }">
          <a-tag :color="ticketStatusColor(record.status)" size="small">{{ ticketStatusText(record.status) }}</a-tag>
        </template>
        <template #action="{ record }">
          <a-link size="small" @click="goTicketDetail(record.id)">查看</a-link>
        </template>
      </a-table>
      <a-empty v-if="projectTickets.length === 0" description="暂无关联工单" />
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/utils/request'
import { getUserList } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { Message, Modal } from '@arco-design/web-vue'

const router = useRouter()
const userStore = useUserStore()
const isAdminOrSupervisor = computed(() => userStore.userInfo?.role === 'admin' || userStore.userInfo?.role === 'supervisor')

const projects = ref<any[]>([])
const users = ref<any[]>([])
const showModal = ref(false)
const editing = ref<any>(null)
const showDrawer = ref(false)
const detailProject = ref<any>(null)
const detailMembers = ref<any[]>([])
const projectTickets = ref<any[]>([])

const defaultForm = () => ({
  name: '',
  description: '',
  type: 'daily',
  priority: 'medium',
  requester: '',
  manager_id: undefined as number | undefined,
  member_ids: [] as number[],
  budget: undefined as number | undefined,
  remark: '',
  start_date: '',
  end_date: '',
  actual_end_date: '',
})

const form = reactive(defaultForm())

const columns = computed(() => {
  const cols = [
    { title: '编号', dataIndex: 'code', width: 110 },
    { title: '项目名称', dataIndex: 'name', ellipsis: true },
    { title: '类型', dataIndex: 'type', slotName: 'type', width: 80 },
    { title: '优先级', dataIndex: 'priority', slotName: 'priority', width: 70 },
    { title: '状态', dataIndex: 'status', slotName: 'status', width: 80 },
    { title: '负责人', dataIndex: 'manager_id', slotName: 'manager', width: 80 },
    { title: '需求方', dataIndex: 'requester', width: 100, ellipsis: true },
    { title: '开始日期', dataIndex: 'start_date', width: 110 },
    { title: '结束日期', dataIndex: 'end_date', width: 110 },
  ]
  if (isAdminOrSupervisor.value) {
    cols.push({ title: '操作', slotName: 'action', width: 160 } as any)
  }
  return cols
})

const statusColor = (s: string) => ({ active: 'green', completed: 'blue', suspended: 'orange' }[s] || 'gray')
const statusText = (s: string) => ({ active: '进行中', completed: '已结束', suspended: '已暂停' }[s] || s)
const typeText = (t: string) => ({ daily: '日常运维', special: '专项任务', emergency: '应急响应', patrol: '巡检项目' }[t] || t)
const priorityColor = (p: string) => ({ high: 'red', medium: 'orange', low: 'blue' }[p] || 'gray')
const priorityText = (p: string) => ({ high: '高', medium: '中', low: '低' }[p] || p)

const ticketColumns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '标题', dataIndex: 'title', ellipsis: true },
  { title: '类型', dataIndex: 'type', slotName: 'ticketType', width: 70 },
  { title: '优先级', dataIndex: 'priority', slotName: 'priority', width: 70 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 80 },
  { title: '操作', slotName: 'action', width: 60 },
]

const ticketPriorityColor = (p: string) => ({ p0: 'red', p1: 'orange', p2: 'gold', p3: 'blue' }[p] || 'gray')
const ticketPriorityText = (p: string) => ({ p0: '紧急', p1: '重大', p2: '严重', p3: '普通' }[p] || p)
const ticketTypeText = (t: string) => ({ fault: '故障', implement: '实施', patrol: '巡检' }[t] || t)
const ticketStatusColor = (s: string) => ({
  created: 'gray', assigned: 'blue', processing: 'cyan',
  suspended: 'orange', review: 'purple', completed: 'green', archived: 'gray'
}[s] || 'gray')
const ticketStatusText = (s: string) => ({
  created: '待派发', assigned: '已派发', processing: '处理中',
  suspended: '已挂起', review: '待验收', completed: '已完单', archived: '已归档'
}[s] || s)

const goTicketDetail = (id: number) => {
  showDrawer.value = false
  router.push(`/tickets/${id}`)
}

const openDetail = async (p: any) => {
  detailProject.value = p
  detailMembers.value = []
  projectTickets.value = []
  showDrawer.value = true

  try {
    const result = await request.get(`/projects/${p.id}`) as any
    detailProject.value = result.project || p
    const memberIDs = result.member_ids || []
    detailMembers.value = users.value.filter(u => memberIDs.includes(u.id))
  } catch (e) {}

  try {
    const result = await request.get('/tickets', { params: { project_id: p.id, page_size: 100 } }) as any
    projectTickets.value = result?.list || []
  } catch (e) {}
}

const getManagerName = (id: number) => {
  const user = users.value.find(u => u.id === id)
  return user ? (user.real_name || user.username) : id
}

const fetchProjects = async () => {
  projects.value = (await request.get('/projects') as any) || []
}

const fetchUsers = async () => {
  try {
    const result = await getUserList({ page: 1, page_size: 200 }) as any
    users.value = result?.list || []
  } catch (e) {}
}

const openCreate = () => {
  editing.value = null
  Object.assign(form, defaultForm())
  showModal.value = true
}

const editProject = async (p: any) => {
  editing.value = p
  try {
    const result = await request.get(`/projects/${p.id}`) as any
    const project = result.project
    const memberIDs = result.member_ids || []
    Object.assign(form, {
      name: project.name,
      description: project.description,
      type: project.type || 'daily',
      priority: project.priority || 'medium',
      requester: project.requester || '',
      manager_id: project.manager_id,
      member_ids: memberIDs,
      budget: project.budget,
      remark: project.remark || '',
      start_date: project.start_date,
      end_date: project.end_date,
      actual_end_date: project.actual_end_date,
    })
  } catch (e) {
    Object.assign(form, {
      name: p.name, description: p.description, type: p.type || 'daily',
      priority: p.priority || 'medium', requester: p.requester || '',
      manager_id: p.manager_id, member_ids: [], budget: p.budget,
      remark: p.remark || '', start_date: p.start_date, end_date: p.end_date,
      actual_end_date: p.actual_end_date,
    })
  }
  showModal.value = true
}

const resetForm = () => {
  editing.value = null
  Object.assign(form, defaultForm())
}

const handleSubmit = async () => {
  if (!form.name) { Message.warning('请输入项目名称'); return }
  if (!form.manager_id) { Message.warning('请选择负责人'); return }
  try {
    if (editing.value) {
      await request.put(`/projects/${editing.value.id}`, form)
      Message.success('更新成功')
    } else {
      await request.post('/projects', form)
      Message.success('创建成功')
    }
    showModal.value = false
    resetForm()
    fetchProjects()
  } catch (e) {}
}

const handleDelete = (id: number) => {
  Modal.confirm({
    title: '确认删除',
    content: '删除项目将同时移除所有成员关联，确定要删除吗？',
    onOk: async () => {
      await request.delete(`/projects/${id}`)
      Message.success('删除成功')
      fetchProjects()
    },
  })
}

onMounted(() => {
  fetchProjects()
  fetchUsers()
})
</script>

<style scoped>
.project-page { padding: 16px; }
</style>
