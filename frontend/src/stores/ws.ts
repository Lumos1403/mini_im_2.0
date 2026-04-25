import { defineStore } from 'pinia'

import type { Message } from '../api/conversation'
import { useAuthStore } from './auth'
import { useChatStore, type OutgoingChatMessage } from './chat'

interface Envelope<T = unknown> {
  seq: string
  type: string
  data: T
  timestamp: number
}

interface WsState {
  connected: boolean
  errorMessage: string
}

interface AckData {
  client_msg_id: string
  message_id: string
  conversation_id: string
  send_status: string
  server_time: string
}

interface FailedData {
  client_msg_id: string
  message_id?: string
  conversation_id: string
  send_status: string
  code: string
  message: string
  server_time?: string
}

interface RecalledData {
  message_id: string
  conversation_id: string
  recalled_by: string
  recalled_at: string
}

let socket: WebSocket | null = null

export const useWsStore = defineStore('ws', {
  state: (): WsState => ({
    connected: false,
    errorMessage: '',
  }),
  actions: {
    connect() {
      const auth = useAuthStore()
      const token = auth.accessToken || localStorage.getItem('access_token')
      if (!token || socket) {
        return
      }

      socket = new WebSocket(buildWebSocketURL(token))
      socket.addEventListener('open', () => {
        this.connected = true
        this.errorMessage = ''
      })
      socket.addEventListener('close', () => {
        this.connected = false
        socket = null
      })
      socket.addEventListener('error', () => {
        this.errorMessage = 'WebSocket 连接异常'
      })
      socket.addEventListener('message', (event) => {
        this.handleEnvelope(event.data)
      })
    },

    disconnect() {
      socket?.close()
      socket = null
      this.connected = false
    },

    sendChatMessage(data: OutgoingChatMessage) {
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        useChatStore().errorMessage = 'WebSocket 未连接，暂不能发送消息'
        return false
      }

      socket.send(
        JSON.stringify({
          seq: crypto.randomUUID(),
          type: 'chat.message.send',
          data,
          timestamp: Date.now(),
        }),
      )
      return true
    },

    handleEnvelope(raw: string) {
      let envelope: Envelope
      try {
        envelope = JSON.parse(raw) as Envelope
      } catch {
        return
      }

      const chat = useChatStore()
      if (envelope.type === 'chat.message.ack') {
        chat.applyAck(envelope.data as AckData)
        return
      }
      if (envelope.type === 'chat.message.failed') {
        chat.applyFailed(envelope.data as FailedData)
        return
      }
      if (envelope.type === 'chat.message.receive') {
        void chat.applyReceive(envelope.data as Message)
        return
      }
      if (envelope.type === 'chat.message.recalled') {
        chat.applyRecalled(envelope.data as RecalledData)
      }
    },
  },
})

function buildWebSocketURL(token: string) {
  const explicitURL = import.meta.env.VITE_WS_URL
  if (explicitURL) {
    const separator = explicitURL.includes('?') ? '&' : '?'
    return `${explicitURL}${separator}token=${encodeURIComponent(token)}`
  }

  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const host = import.meta.env.DEV ? 'localhost:8081' : window.location.host
  return `${protocol}://${host}/ws?token=${encodeURIComponent(token)}`
}
