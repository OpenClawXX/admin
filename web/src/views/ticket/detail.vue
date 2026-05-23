<template>
  <div class="ticket-detail">
    <a-page-header title="工单详情" @back="router.back()" />
    <a-card :bordered="false" v-if="ticket">
      <a-descriptions :column="2" bordered>
        <a-descriptions-item label="工单ID">{{ ticket.id }}</a-descriptions-item>
        <a-descriptions-item label="标题">{{ ticket.title }}</a-descriptions-item>
        <a-descriptions-item label="类型">{{ typeText(ticket.type) }}</a-descriptions-item>
        <a-descriptions-item label="优先级">
          <a-tag :color="priorityColor(ticket.priority)">{{ priorityText(ticket.priority) }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag :color="statusColor(ticket.status)">{{ statusText(ticket.status) }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">{{ ticket.created_at }}</a-descriptions-item>
        <a-descriptions-item label="所属项目">
          <a-link v-if="projectInfo" @click="router.push('/projects')">{{ projectInfo.code }} - {{ projectInfo.name }}</a-link>
          <span v-else>-</span>
        </a-descriptions-item>
        <a-descriptions-item label="描述" :span="2">{{ ticket.description }}</a-descriptions-item>
      </a-descriptions>

      <a-divider>操作</a-divider>
      <a-space>
        <a-button v-if="canAssign" type="primary" @click="showAssignModal = true">派单</a-button>
        <a-button v-if="canTransfer" @click="showTransferModal = true">转派</a-button>
        <a-button v-if="canSuspend" status="warning" @click="handleSuspend">挂起</a-button>
        <a-button v-if="canResume" status="success" @click="handleResume">恢复</a-button>
        <a-button v-if="canComplete" type="primary" @click="openCompleteModal">完单</a-button>
        <a-button v-if="canReview" status="success" @click="handleReview(true)">通过</a-button>
        <a-button v-if="canReview" status="danger" @click="handleReview(false)">驳回</a-button>
        <a-button v-if="canArchive" @click="handleArchive">归档</a-button>
        <a-button v-if="isAdmin" status="danger" @click="handleDelete">删除工单</a-button>
      </a-space>

      <!-- Completion Report -->
      <template v-if="completion">
        <a-divider>完单报告</a-divider>
        <a-descriptions :column="2" bordered size="small">
          <a-descriptions-item label="处理结果" :span="2">
            <a-tag :color="resultColor(completion.result)">{{ resultText(completion.result) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="解决方案" :span="2">{{ completion.solution }}</a-descriptions-item>
          <a-descriptions-item label="根因分析" :span="2">{{ completion.root_cause || '-' }}</a-descriptions-item>
          <a-descriptions-item label="影响范围">{{ completion.impact || '-' }}</a-descriptions-item>
          <a-descriptions-item label="遗留问题">{{ completion.remaining || '-' }}</a-descriptions-item>
          <a-descriptions-item label="后续建议" :span="2">{{ completion.suggestion || '-' }}</a-descriptions-item>
          <a-descriptions-item label="交接备注" :span="2">{{ completion.handover || '-' }}</a-descriptions-item>
        </a-descriptions>
        <div v-if="ticketFiles.length > 0" style="margin-top: 12px">
          <strong>附件：</strong>
          <a-space direction="vertical" style="margin-top: 8px; width: 100%">
            <div v-for="f in ticketFiles" :key="f.id" style="display: flex; align-items: center; gap: 8px">
              <icon-file />
              <span>{{ f.filename }}</span>
              <span style="color: #86909c; font-size: 12px">({{ formatSize(f.filesize) }}, {{ getUserName(f.uploader_id) }})</span>
              <a-link @click="downloadFile(f)">下载</a-link>
              <a-link v-if="canDeleteFile(f)" status="danger" @click="handleDeleteFile(f)">删除</a-link>
            </div>
          </a-space>
        </div>
      </template>

      <a-divider>进度上报</a-divider>
      <a-space style="width: 100%">
        <a-textarea v-model="progressContent" placeholder="输入进度内容" style="flex: 1" />
        <a-button type="primary" @click="handleAddProgress">提交进度</a-button>
      </a-space>

      <a-divider>流转日志</a-divider>
      <a-timeline>
        <a-timeline-item v-for="log in logs" :key="log.id">
          <p>{{ log.content }}</p>
          <p style="color: #86909c; font-size: 12px">
            <span>{{ getUserName(log.operator_id) }}</span>
            <span style="margin-left: 8px">{{ log.created_at }}</span>
          </p>
        </a-timeline-item>
      </a-timeline>
    </a-card>

    <!-- Assign Modal -->
    <a-modal v-model:visible="showAssignModal" title="派单" @ok="handleAssign">
      <a-form-item label="选择工程师">
        <a-select v-model="assigneeId" placeholder="选择要指派的工程师">
          <a-option v-for="eng in engineers" :key="eng.id" :value="eng.id">
            {{ eng.real_name || eng.username }}
          </a-option>
        </a-select>
      </a-form-item>
      <a-form-item label="备注">
        <a-input v-model="assignRemark" />
      </a-form-item>
    </a-modal>

    <!-- Transfer Modal -->
    <a-modal v-model:visible="showTransferModal" title="转派" @ok="handleTransfer">
      <a-form-item label="选择工程师">
        <a-select v-model="transferId" placeholder="选择要转派的工程师">
          <a-option v-for="eng in engineers" :key="eng.id" :value="eng.id">
            {{ eng.real_name || eng.username }}
          </a-option>
        </a-select>
      </a-form-item>
      <a-form-item label="备注">
        <a-input v-model="transferRemark" />
      </a-form-item>
    </a-modal>

    <!-- Complete Modal -->
    <a-modal v-model:visible="showCompleteModal" title="完单提交" @ok="handleComplete" :width="640">
      <a-form :model="completionForm" layout="vertical">
        <a-form-item field="solution" label="解决方案" :rules="[{ required: true, message: '请填写解决方案' }]">
          <a-textarea v-model="completionForm.solution" placeholder="详细描述处理过程和最终方案" :max-length="5000" show-word-limit />
        </a-form-item>
        <a-form-item field="root_cause" label="根因分析">
          <a-textarea v-model="completionForm.root_cause" placeholder="问题根本原因分析" :max-length="2000" show-word-limit />
        </a-form-item>
        <a-form-item field="result" label="处理结果">
          <a-select v-model="completionForm.result">
            <a-option value="resolved">完全解决</a-option>
            <a-option value="partial">部分解决</a-option>
            <a-option value="escalate">无法解决需升级</a-option>
          </a-select>
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="impact" label="影响范围">
              <a-textarea v-model="completionForm.impact" placeholder="本次处理影响到的系统/服务范围" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="remaining" label="遗留问题">
              <a-textarea v-model="completionForm.remaining" placeholder="未解决的遗留事项" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="suggestion" label="后续建议">
              <a-textarea v-model="completionForm.suggestion" placeholder="预防措施或改进建议" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="handover" label="交接备注">
              <a-textarea v-model="completionForm.handover" placeholder="给接手人的补充说明" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="附件上传">
          <a-upload
            :action="`/api/tickets/${ticketId}/files`"
            :headers="uploadHeaders"
            :file-list="existingFiles"
            multiple
            auto-upload
            ref="uploadRef"
            @change="onUploadChange"
          >
            <template #upload-button>
              <div style="padding: 16px; border: 1px dashed #c9cdd4; border-radius: 4px; text-align: center; cursor: pointer">
                <icon-upload /> 点击或拖拽上传（截图/日志/配置文件等）
              </div>
            </template>
          </a-upload>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getTicketDetail, getTicketLogs, assignTicket, transferTicket,
  suspendTicket, resumeTicket, addProgress, completeTicket,
  reviewTicket, archiveTicket, deleteTicket
} from '@/api/ticket'
import { getUserList } from '@/api/user'
import request from '@/utils/request'
import { useUserStore } from '@/stores/user'
import { Message, Modal } from '@arco-design/web-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const ticket = ref<any>(null)
const logs = ref<any[]>([])
const engineers = ref<any[]>([])
const allUsers = ref<any[]>([])
const projectInfo = ref<any>(null)
const completion = ref<any>(null)
const ticketFiles = ref<any[]>([])
const progressContent = ref('')
const showAssignModal = ref(false)
const showTransferModal = ref(false)
const showCompleteModal = ref(false)
const assigneeId = ref<number>()
const assignRemark = ref('')
const transferId = ref<number>()
const transferRemark = ref('')
const uploadRef = ref<any>(null)
const uploading = ref(false)
const existingFiles = ref<any[]>([])

const completionForm = reactive({
  solution: '',
  root_cause: '',
  result: 'resolved',
  impact: '',
  remaining: '',
  suggestion: '',
  handover: '',
})

const ticketId = computed(() => Number(route.params.id))
const isAdmin = computed(() => userStore.userInfo?.role === 'admin')
const isSupervisor = computed(() => userStore.userInfo?.role === 'supervisor')

const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${userStore.token}`,
}))

const canAssign = computed(() => (isAdmin.value || isSupervisor.value) && ['created', 'pending'].includes(ticket.value?.status))
const canTransfer = computed(() => (isAdmin.value || isSupervisor.value) && !['completed', 'archived'].includes(ticket.value?.status))
const canSuspend = computed(() => ['assigned', 'processing'].includes(ticket.value?.status))
const canResume = computed(() => ticket.value?.status === 'suspended')
const canComplete = computed(() => ['processing', 'assigned'].includes(ticket.value?.status))
const canReview = computed(() => (isAdmin.value || isSupervisor.value) && ticket.value?.status === 'review')
const canArchive = computed(() => (isAdmin.value || isSupervisor.value) && ticket.value?.status === 'completed')

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
const resultColor = (r: string) => ({ resolved: 'green', partial: 'orange', escalate: 'red' }[r] || 'gray')
const resultText = (r: string) => ({ resolved: '完全解决', partial: '部分解决', escalate: '无法解决需升级' }[r] || r)

const formatSize = (bytes: number) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

const getUserName = (id: number) => {
  const user = allUsers.value.find(u => u.id === id)
  return user ? (user.real_name || user.username) : `ID:${id}`
}

const fetchAllUsers = async () => {
  try {
    const result = await getUserList({ page: 1, page_size: 200 }) as any
    allUsers.value = result?.list || []
  } catch (e) {}
}

const fetchCompletion = async () => {
  try {
    completion.value = await request.get(`/tickets/${ticketId.value}/completion`)
  } catch (e: any) {
    completion.value = null
  }
}

const fetchFiles = async () => {
  try {
    const result = await request.get(`/tickets/${ticketId.value}/files`) as any
    ticketFiles.value = Array.isArray(result) ? result : []
  } catch (e: any) {
    ticketFiles.value = []
  }
}

const fetchData = async () => {
  ticket.value = await getTicketDetail(ticketId.value)
  logs.value = await getTicketLogs(ticketId.value) as any || []
  projectInfo.value = null
  if (ticket.value?.project_id) {
    try {
      const result = await request.get(`/projects/${ticket.value.project_id}`) as any
      projectInfo.value = result?.project || null
    } catch (e) {}
  }
  fetchCompletion()
  fetchFiles()
}

const openCompleteModal = async () => {
  if (completion.value) {
    Object.assign(completionForm, {
      solution: completion.value.solution || '',
      root_cause: completion.value.root_cause || '',
      result: completion.value.result || 'resolved',
      impact: completion.value.impact || '',
      remaining: completion.value.remaining || '',
      suggestion: completion.value.suggestion || '',
      handover: completion.value.handover || '',
    })
  } else {
    Object.assign(completionForm, {
      solution: '', root_cause: '', result: 'resolved',
      impact: '', remaining: '', suggestion: '', handover: '',
    })
  }
  // Load existing files into upload list
  existingFiles.value = []
  try {
    const result = await request.get(`/tickets/${ticketId.value}/files`) as any
    const files = Array.isArray(result) ? result : []
    existingFiles.value = files.map((f: any) => ({
      uid: `existing_${f.id}`,
      name: `${f.filename} (${getUserName(f.uploader_id)})`,
      status: 'done',
      url: `/api/tickets/${ticketId.value}/files/${f.id}/download`,
      response: { data: { id: f.id, uploader_id: f.uploader_id } },
    }))
  } catch (e) {}
  showCompleteModal.value = true
}

const onUploadChange = async ({ fileList, file }: any) => {
  uploading.value = fileList.some((f: any) => f.status === 'uploading')
  // When a file is removed, delete from server if it was uploaded and user has permission
  if (file.status === 'removed' && file.response?.data?.id) {
    const fileUploaderId = file.response.data.uploader_id
    const canDelete = isAdmin.value || isSupervisor.value || fileUploaderId === userStore.userInfo?.id
    if (canDelete) {
      try {
        await request.delete(`/tickets/${ticketId.value}/files/${file.response.data.id}`)
      } catch (e) {}
    }
  }
}

const downloadFile = async (f: any) => {
  try {
    const response = await fetch(`/api/tickets/${ticketId.value}/files/${f.id}/download`, {
      headers: { Authorization: `Bearer ${userStore.token}` },
    })
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = f.filename
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    Message.error('下载失败')
  }
}

const canDeleteFile = (f: any) => {
  if (isAdmin.value || isSupervisor.value) return true
  return f.uploader_id === userStore.userInfo?.id
}

const handleDeleteFile = async (f: any) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除文件 "${f.filename}" 吗？`,
    onOk: async () => {
      try {
        await request.delete(`/tickets/${ticketId.value}/files/${f.id}`)
        Message.success('文件已删除')
        fetchFiles()
      } catch (e) {
        Message.error('删除失败')
      }
    },
  })
}

const handleAssign = async () => {
  if (!assigneeId.value) return
  await assignTicket(ticketId.value, { assignee_id: assigneeId.value, remark: assignRemark.value })
  Message.success('派单成功')
  showAssignModal.value = false
  fetchData()
}

const handleTransfer = async () => {
  if (!transferId.value) return
  await transferTicket(ticketId.value, { assignee_id: transferId.value, remark: transferRemark.value })
  Message.success('转派成功')
  showTransferModal.value = false
  fetchData()
}

const handleSuspend = async () => {
  await suspendTicket(ticketId.value)
  Message.success('已挂起')
  fetchData()
}

const handleResume = async () => {
  await resumeTicket(ticketId.value)
  Message.success('已恢复')
  fetchData()
}

const handleAddProgress = async () => {
  if (!progressContent.value) return
  await addProgress(ticketId.value, { content: progressContent.value, status: 'processing' })
  Message.success('进度已上报')
  progressContent.value = ''
  fetchData()
}

const handleComplete = async () => {
  if (!completionForm.solution) {
    Message.warning('请填写解决方案')
    return
  }
  if (uploading.value) {
    Message.warning('文件上传中，请等待上传完成')
    return
  }
  try {
    // 1. Submit completion report
    await request.post(`/tickets/${ticketId.value}/completion`, completionForm)
    // 2. Complete the ticket
    await completeTicket(ticketId.value, { solution: completionForm.solution })
    Message.success('完单报告已提交')
    showCompleteModal.value = false
    fetchData()
  } catch (e) {
    Message.error('提交失败')
  }
}

const handleReview = async (approved: boolean) => {
  await reviewTicket(ticketId.value, { approved })
  Message.success(approved ? '验收通过' : '已驳回')
  fetchData()
}

const handleArchive = async () => {
  await archiveTicket(ticketId.value)
  Message.success('已归档')
  fetchData()
}

const handleDelete = () => {
  Modal.confirm({
    title: '确认删除',
    content: '删除后不可恢复，确定要删除该工单吗？',
    onOk: async () => {
      await deleteTicket(ticketId.value)
      Message.success('工单已删除')
      router.push('/tickets')
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

onMounted(() => {
  fetchData()
  fetchEngineers()
  fetchAllUsers()
})
</script>

<style scoped>
.ticket-detail {
  padding: 16px;
}
</style>
