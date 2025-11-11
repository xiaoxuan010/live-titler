export interface Lyric {
  transition_time: string | number
  text: string
}

export interface PresetContent {
  num?: string
  person?: string
  name?: string
  song_name?: string
  current_lyrics?: number
  lyrics?: Lyric[]
}

export interface Preset {
  key_num: string
  key_name: string
  status: string
  transition_time: string
  current_preset: string
  content: PresetContent[]
}

export type PresetStatus = 'CLOSED' | 'CLOSING' | 'OPENED' | 'OPENING' | 'PLAYING_FORWARD'

// New normalized data types for key management
export type KeyType = 'program' | 'lyrics'

export interface LyricData {
  id: number
  text: string
  transition_time: number
  position: number
}

export interface SongData {
  id: number
  name: string
  current_lyric_id: number | null
  position: number
  lyrics: LyricData[]
}

export interface ProgramData {
  id: number
  num: string
  name: string
  person: string
  position: number
}

export interface Key {
  id: number
  name: string
  position: number
  key_type: KeyType
  status: PresetStatus
  transition_time: number
  current_preset_id: number | null
  version: number
  programs?: ProgramData[]
  songs?: SongData[]
}
