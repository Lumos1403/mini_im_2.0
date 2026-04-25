import axios from 'axios'

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 10000,
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export function unwrap<T>(response: ApiResponse<T>): T {
  if (response.code !== 0) {
    throw new Error(response.message)
  }
  return response.data
}

export function getApiErrorMessage(error: unknown): string {
  if (axios.isAxiosError<ApiResponse<unknown>>(error)) {
    const message = error.response?.data?.message
    if (message) {
      return message
    }
  }

  return error instanceof Error ? error.message : '请求失败'
}
