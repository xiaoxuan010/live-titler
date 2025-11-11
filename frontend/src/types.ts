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
  programs?: ProgramData[];
  songs?: SongData[];
}

// Type guards
export function isProgramKey(
  key: Key,
): key is Key & { programs: ProgramData[] } {
  return key.key_type === "program" && !!key.programs;
}

export function isLyricsKey(key: Key): key is Key & { songs: SongData[] } {
  return key.key_type === "lyrics" && !!key.songs;
}
