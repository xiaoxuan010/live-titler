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
