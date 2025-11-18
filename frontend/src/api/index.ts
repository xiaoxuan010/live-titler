import type { Key, KeyType, ProgramData, SongData, LyricData } from "../types";

export class ApiConflictError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ApiConflictError";
  }
}

const API_BASE = "/api";

export const api = {
  // Keys management
  async getKeys(): Promise<Key[]> {
    const response = await fetch(`${API_BASE}/keys`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  },

  async createKey(
    name: string,
    keyType: KeyType,
    transitionTime: number = 1.0,
  ): Promise<Key> {
    const response = await fetch(`${API_BASE}/keys`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name,
        key_type: keyType,
        transition_time: transitionTime,
      }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  },

  async updateKey(
    id: number,
    data: Partial<Key> & { version: number },
  ): Promise<{ status: string; version: number }> {
    const response = await fetch(`${API_BASE}/keys/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      if (response.status === 409) {
        throw new ApiConflictError("控制版本冲突，请重试");
      }
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  },

  async deleteKey(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/keys/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  async reorderKeys(ids: number[]): Promise<void> {
    const response = await fetch(`${API_BASE}/keys/reorder`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ ids }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  // Programs management
  async createProgram(
    keyId: number,
    data: Omit<ProgramData, "id" | "position">,
  ): Promise<ProgramData> {
    const response = await fetch(`${API_BASE}/programs`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ key_id: keyId, ...data }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  },

  async updateProgram(
    id: number,
    data: Partial<Omit<ProgramData, "id" | "position">>,
  ): Promise<void> {
    const response = await fetch(`${API_BASE}/programs/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  async deleteProgram(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/programs/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  // Songs management
  async createSong(keyId: number, name: string): Promise<SongData> {
    const response = await fetch(`${API_BASE}/songs`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ key_id: keyId, name }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  },

  async updateSong(
    id: number,
    data: Partial<Omit<SongData, "id" | "position" | "lyrics">>,
  ): Promise<void> {
    const response = await fetch(`${API_BASE}/songs/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  async deleteSong(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/songs/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  // Lyrics management
  async createLyric(
    songId: number,
    text: string,
    transitionTime: number = 1.0,
  ): Promise<LyricData> {
    const response = await fetch(`${API_BASE}/lyrics`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        song_id: songId,
        text,
        transition_time: transitionTime,
      }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  },

  async updateLyric(
    id: number,
    data: Partial<Omit<LyricData, "id" | "position">>,
  ): Promise<void> {
    const response = await fetch(`${API_BASE}/lyrics/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  async deleteLyric(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/lyrics/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  // Reordering
  async reorderPrograms(ids: number[]): Promise<void> {
    const response = await fetch(`${API_BASE}/programs/reorder`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ ids }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  async reorderLyrics(ids: number[]): Promise<void> {
    const response = await fetch(`${API_BASE}/lyrics/reorder`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ ids }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  // CSV Import/Export for Programs
  async importProgramsCSV(keyId: number, csv: string): Promise<void> {
    const response = await fetch(`${API_BASE}/keys/${keyId}/import-csv`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ csv }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  async exportProgramsCSV(keyId: number): Promise<string> {
    const response = await fetch(`${API_BASE}/keys/${keyId}/export-csv`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    return data.csv;
  },

  // CSV Import/Export for Lyrics
  async importLyricsCSV(songId: number, csv: string): Promise<void> {
    const response = await fetch(`${API_BASE}/songs/${songId}/import-csv`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ csv }),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  async exportLyricsCSV(songId: number): Promise<string> {
    const response = await fetch(`${API_BASE}/songs/${songId}/export-csv`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    return data.csv;
  },

  // JSON Import/Export for Key
  async importKeyJSON(keyId: number, data: Key): Promise<void> {
    const response = await fetch(`${API_BASE}/keys/${keyId}/import-json`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  },

  async exportKeyJSON(keyId: number): Promise<Key> {
    const response = await fetch(`${API_BASE}/keys/${keyId}/export-json`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  },
};

export function createWebSocket(onMessage: (data: Key[]) => void): WebSocket {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsUrl = `${protocol}//${window.location.host}/ws`;

  const ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    console.log("WebSocket connected");
  };

  ws.onmessage = (event) => {
    try {
      const data: Key[] = JSON.parse(event.data);
      onMessage(data);
    } catch (err) {
      console.error("Failed to parse WebSocket message:", err);
    }
  };

  ws.onerror = (error) => {
    console.error("WebSocket error:", error);
  };

  return ws;
}
