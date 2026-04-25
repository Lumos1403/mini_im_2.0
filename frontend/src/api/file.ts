import { http, unwrap, type ApiResponse } from './http'

export interface FileUploadResult {
  file_id: string
  original_name: string
  file_size: number
  mime_type: string
}

export interface FileDownloadResult {
  blob: Blob
  fileName: string
}

export async function uploadFile(file: File): Promise<FileUploadResult> {
  const formData = new FormData()
  formData.append('file', file)

  const { data } = await http.post<ApiResponse<FileUploadResult>>('/api/files/upload', formData)
  return unwrap(data)
}

export async function downloadFile(fileID: string): Promise<FileDownloadResult> {
  const response = await http.get<Blob>(`/api/files/${encodeURIComponent(fileID)}/download`, {
    responseType: 'blob',
  })

  return {
    blob: response.data,
    fileName: parseContentDispositionFileName(response.headers['content-disposition']),
  }
}

function parseContentDispositionFileName(disposition: unknown) {
  if (typeof disposition !== 'string' || !disposition.trim()) {
    return ''
  }

  const encodedMatch = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (encodedMatch?.[1]) {
    try {
      return decodeURIComponent(encodedMatch[1].trim())
    } catch {
      return encodedMatch[1].trim()
    }
  }

  const quotedMatch = disposition.match(/filename="([^"]+)"/i)
  if (quotedMatch?.[1]) {
    return quotedMatch[1].trim()
  }

  const plainMatch = disposition.match(/filename=([^;]+)/i)
  return plainMatch?.[1]?.trim() || ''
}
