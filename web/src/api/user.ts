import request from '@/utils/request'

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  user: UserInfo
}

export interface UserInfo {
  id: number
  username: string
  real_name: string
  email: string
  phone: string
  role: string
  team_id: number | null
  skills: string[]
  status: number
}

export const login = (data: LoginParams): Promise<LoginResult> =>
  request.post('/login', data)

export const getProfile = (): Promise<UserInfo> =>
  request.get('/profile')

export const getUserList = (params?: { page?: number; page_size?: number; team_id?: number }) =>
  request.get('/users', { params })

export const createUser = (data: any) =>
  request.post('/users', data)

export const updateUser = (id: number, data: any) =>
  request.put(`/users/${id}`, data)

export const deleteUser = (id: number) =>
  request.delete(`/users/${id}`)

export const resetPassword = (id: number, password?: string): Promise<{ password: string }> =>
  request.post(`/users/${id}/reset-password`, { password: password || '' })

export const getTeamList = () =>
  request.get('/teams')

export const createTeam = (data: any) =>
  request.post('/teams', data)

export const updateTeam = (id: number, data: any) =>
  request.put(`/teams/${id}`, data)

export const deleteTeam = (id: number) =>
  request.delete(`/teams/${id}`)
