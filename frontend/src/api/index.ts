import type { Preset, Key, KeyType } from '../types'

const API_BASE = '/api'

export const api = {
  async getPresets(): Promise<Preset[]> {
    const response = await fetch(`${API_BASE}/presets`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },

  async updatePreset(keyNum: string, preset: Preset): Promise<void> {
    const response = await fetch(`${API_BASE}/presets/${keyNum}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(preset),
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
  },

  async importPresets(presets: Preset[]): Promise<void> {
    const response = await fetch(`${API_BASE}/presets/import`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(presets),
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
  },

  async exportPresets(): Promise<Preset[]> {
    const response = await fetch(`${API_BASE}/presets/export`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },

  // New normalized API methods for key management
  async getKeys(): Promise<Key[]> {
    const response = await fetch(`${API_BASE}/keys`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },

  async createKey(name: string, keyType: KeyType): Promise<Key> {
    const response = await fetch(`${API_BASE}/keys`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name, key_type: keyType }),
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },

  async updateKey(id: number, data: Partial<Key>): Promise<Key> {
    const response = await fetch(`${API_BASE}/keys/${id}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    if (!response.ok) {
      if (response.status === 409) {
        throw new Error('版本冲突：数据已被其他客户端修改，请刷新后重试')
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },

  async deleteKey(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/keys/${id}`, {
      method: 'DELETE',
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
  },

  async reorderKeys(ids: number[]): Promise<void> {
    const response = await fetch(`${API_BASE}/keys/reorder`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ ids }),
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
  },
}

export function createWebSocket(onMessage: (data: Preset[]) => void): WebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws`
  
  const ws = new WebSocket(wsUrl)
  
  ws.onopen = () => {
    console.log('WebSocket connected')
  }
  
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      onMessage(data)
    } catch (err) {
      console.error('Failed to parse WebSocket message:', err)
    }
  }
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
  }
  
  return ws
}
