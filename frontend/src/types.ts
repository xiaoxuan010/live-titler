// New normalized data types for key management
export type KeyType = "program" | "lyrics";
export type PresetStatus =
  | "CLOSED"
  | "CLOSING"
  | "OPENED"
  | "OPENING"
  | "PLAYING_FORWARD";

export interface LyricData {
  id: number;
  text: string;
  transition_time: number;
  position: number;
}

export interface SongData {
  id: number;
  name: string;
  current_lyric_id: number | null;
  position: number;
  lyrics: LyricData[];
}

export interface ProgramData {
  id: number;
  num: string;
  name: string;
  person: string;
  position: number;
}

export interface Key {
  id: number;
  name: string;
  position: number;
  key_type: KeyType;
  status: PresetStatus;
  transition_time: number;
  current_preset_id: number | null;
  version: number;
  // backend may return `null` for empty relations; allow null here
  programs?: ProgramData[] | null;
  songs?: SongData[] | null;
}

// Type guards
export function isProgramKey(key: Key) {
  return key.key_type === "program";
}

export function isLyricsKey(key: Key) {
  return key.key_type === "lyrics";
}

// Safe accessors that normalize possibly-null backend fields to empty arrays.
export function programsOf(key: Key): ProgramData[] {
  return key.programs ?? [];
}

export function songsOf(key: Key): SongData[] {
  return key.songs ?? [];
}
