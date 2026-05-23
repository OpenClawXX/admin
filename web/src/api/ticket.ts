import request from '@/utils/request'

export interface TicketQuery {
  status?: string
  priority?: string
  type?: string
  assignee_id?: number
  project_id?: number
  keyword?: string
  page?: number
  page_size?: number
}

export interface TicketInfo {
  id: number
  title: string
  description: string
  type: string
  priority: string
  status: string
  creator_id: number
  assignee_id: number | null
  project_id: number | null
  sla_deadline: string | null
  created_at: string
  updated_at: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export const getTicketList = (params?: TicketQuery): Promise<PageResult<TicketInfo>> =>
  request.get('/tickets', { params })

export const getTicketDetail = (id: number): Promise<TicketInfo> =>
  request.get(`/tickets/${id}`)

export const createTicket = (data: any) =>
  request.post('/tickets', data)

export const updateTicket = (id: number, data: any) =>
  request.put(`/tickets/${id}`, data)

export const assignTicket = (id: number, data: { assignee_id: number; remark?: string }) =>
  request.post(`/tickets/${id}/assign`, data)

export const transferTicket = (id: number, data: { assignee_id: number; remark?: string }) =>
  request.post(`/tickets/${id}/transfer`, data)

export const suspendTicket = (id: number, data?: { reason?: string }) =>
  request.post(`/tickets/${id}/suspend`, data)

export const resumeTicket = (id: number) =>
  request.post(`/tickets/${id}/resume`)

export const addProgress = (id: number, data: { content: string; status?: string }) =>
  request.post(`/tickets/${id}/progress`, data)

export const addTicketLog = (id: number, data: { content: string }) =>
  request.post(`/tickets/${id}/logs`, data)

export const completeTicket = (id: number, data: { solution: string; remark?: string }) =>
  request.post(`/tickets/${id}/complete`, data)

export const reviewTicket = (id: number, data: { approved: boolean; remark?: string }) =>
  request.post(`/tickets/${id}/review`, data)

export const archiveTicket = (id: number) =>
  request.post(`/tickets/${id}/archive`)

export const deleteTicket = (id: number) =>
  request.delete(`/tickets/${id}`)

export const getTicketLogs = (id: number) =>
  request.get(`/tickets/${id}/logs`)
