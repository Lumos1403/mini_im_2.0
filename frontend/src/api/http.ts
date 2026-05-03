import axios, { type AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
}

type RetryableConfig = InternalAxiosRequestConfig & { _authRetry?: boolean }

const authExpiredCodes = new Set([20001, 20002])
const refreshInvalidCodes = new Set([20003])

let refreshPromise: Promise<TokenPair> | null = null
let authFailureHandler: (() => void | Promise<void>) | null = null
let tokenUpdateHandler: ((accessToken: string, refreshToken: string) => void) | null = null

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

http.interceptors.response.use(
  async (response) => {
    const apiResponse = response.data as Partial<ApiResponse<unknown>> | undefined
    if (!apiResponse || typeof apiResponse.code !== 'number') {
      return response
    }

    if (refreshInvalidCodes.has(apiResponse.code)) {
      await notifyAuthFailure()
      return response
    }

    if (!authExpiredCodes.has(apiResponse.code) || isAuthEndpoint(response.config.url)) {
      return response
    }

    return retryWithFreshToken(response)
  },
  async (error: AxiosError<ApiResponse<unknown>>) => {
    const status = error.response?.status
    const code = error.response?.data?.code
    const config = error.config as RetryableConfig | undefined
    if (
      config &&
      !config._authRetry &&
      !isAuthEndpoint(config.url) &&
      (status === 401 || code === 20001 || code === 20002)
    ) {
      try {
        const tokens = await refreshAccessToken()
        applyFreshTokens(tokens)
        config._authRetry = true
        config.headers.Authorization = `Bearer ${tokens.access_token}`
        return http.request(config)
      } catch (refreshError) {
        await notifyAuthFailure()
        return Promise.reject(refreshError)
      }
    }

    if (code === 20003) {
      await notifyAuthFailure()
    }
    return Promise.reject(error)
  },
)

export function unwrap<T>(response: ApiResponse<T>): T {
  if (response.code !== 0) {
    throw new Error(response.message)
  }
  return response.data
}

export function setAuthFailureHandler(handler: (() => void | Promise<void>) | null) {
  authFailureHandler = handler
}

export function setTokenUpdateHandler(handler: ((accessToken: string, refreshToken: string) => void) | null) {
  tokenUpdateHandler = handler
}

export function getApiErrorMessage(error: unknown): string {
  if (axios.isAxiosError<ApiResponse<unknown>>(error)) {
    const message = error.response?.data?.message
    if (message) {
      return message
    }
  }

  return error instanceof Error ? error.message : 'Request failed'
}

async function retryWithFreshToken(response: AxiosResponse) {
  const config = response.config as RetryableConfig
  if (config._authRetry) {
    await notifyAuthFailure()
    return response
  }

  try {
    const tokens = await refreshAccessToken()
    applyFreshTokens(tokens)
    config._authRetry = true
    config.headers.Authorization = `Bearer ${tokens.access_token}`
    return http.request(config)
  } catch (error) {
    await notifyAuthFailure()
    return Promise.reject(error)
  }
}

function isAuthEndpoint(url = '') {
  return (
    url.includes('/api/auth/login') ||
    url.includes('/api/auth/register') ||
    url.includes('/api/auth/refresh') ||
    url.includes('/api/auth/logout')
  )
}

function applyFreshTokens(tokens: TokenPair) {
  localStorage.setItem('access_token', tokens.access_token)
  localStorage.setItem('refresh_token', tokens.refresh_token)
  tokenUpdateHandler?.(tokens.access_token, tokens.refresh_token)
}

async function refreshAccessToken() {
  const refreshToken = localStorage.getItem('refresh_token')
  if (!refreshToken) {
    throw new Error('missing refresh token')
  }

  if (!refreshPromise) {
    refreshPromise = axios
      .post<ApiResponse<TokenPair>>(
        `${http.defaults.baseURL || ''}/api/auth/refresh`,
        { refresh_token: refreshToken },
        { timeout: http.defaults.timeout },
      )
      .then((response) => {
        if (response.data.code !== 0) {
          throw new Error(response.data.message)
        }
        return response.data.data
      })
      .finally(() => {
        refreshPromise = null
      })
  }

  return refreshPromise
}

async function notifyAuthFailure() {
  await authFailureHandler?.()
}
