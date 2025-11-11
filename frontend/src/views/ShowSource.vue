<template>
  <div class="fixed inset-0 overflow-hidden bg-transparent">
    <!-- Program Keys Display -->
    <div
      v-for="key in programKeys"
      :key="key.id"
      v-show="key.status === 'OPENED' || key.status === 'OPENING'"
      class="absolute inset-0 flex items-center justify-center"
      :style="{
        transition: key.status === 'OPENING' || key.status === 'CLOSING'
          ? `opacity ${key.transition_time}s ease`
          : 'none',
        opacity: key.status === 'OPENED' || key.status === 'OPENING' ? 1 : 0
      }"
    >
      <div class="text-center px-8" v-if="getCurrentProgram(key)">
        <div class="text-6xl font-bold text-white mb-4" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
          {{ getCurrentProgram(key)!.name }}
        </div>
        <div class="text-4xl text-white" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
          {{ getCurrentProgram(key)!.person }}
        </div>
      </div>
    </div>

    <!-- Lyrics Keys Display -->
    <div
      v-for="key in lyricsKeys"
      :key="key.id"
      class="absolute bottom-20 left-0 right-0 flex justify-center"
      :style="{
        transition: shouldShowLyrics(key)
          ? 'none'
          : `opacity ${key.transition_time}s ease`,
        opacity: shouldShowLyrics(key) ? 1 : 0
      }"
    >
      <div class="text-5xl font-bold text-white px-8 text-center" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
        <span
          v-for="(char, index) in getCurrentLyricText(key)"
          :key="index"
          :ref="el => { if (el) setCharRef(key.id, index, el as HTMLElement) }"
          class="inline-block"
          :style="charStyles[key.id]?.[index] || {}"
        >
          {{ char }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { createWebSocket } from '@/api'
import type { Key, ProgramData, SongData } from '@/types'
import { isProgramKey, isLyricsKey } from '@/types'

const keys = ref<Key[]>([])
const reconnectAttempts = ref(0)
const reconnectDelay = 1000
let ws: WebSocket | null = null

const charRefs = ref<Record<number, Record<number, HTMLElement>>>({})
const charStyles = ref<Record<number, Record<number, any>>>({})

const programKeys = computed(() => keys.value.filter(isProgramKey))
const lyricsKeys = computed(() => keys.value.filter(isLyricsKey))

function setCharRef(keyId: number, index: number, el: HTMLElement) {
  if (!charRefs.value[keyId]) {
    charRefs.value[keyId] = {}
  }
  charRefs.value[keyId][index] = el
}

function connectWebSocket() {
  console.log(`Attempting WebSocket connection (attempt ${reconnectAttempts.value + 1})...`)
  
  try {
    ws = createWebSocket((data: Key[]) => {
      keys.value = data.sort((a, b) => a.position - b.position)
      reconnectAttempts.value = 0
    })
    
    ws.onopen = () => {
      console.log('WebSocket connected successfully')
      reconnectAttempts.value = 0
    }
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
    
    ws.onclose = (event) => {
      console.log(`WebSocket disconnected (code: ${event.code})`)
      reconnectAttempts.value++
      console.log(`Will reconnect in ${reconnectDelay}ms (attempt ${reconnectAttempts.value})...`)
      setTimeout(connectWebSocket, reconnectDelay)
    }
  } catch (err) {
    console.error('Failed to create WebSocket connection:', err)
    reconnectAttempts.value++
    console.log(`Will retry connection in ${reconnectDelay}ms...`)
    setTimeout(connectWebSocket, reconnectDelay)
  }
}

function getCurrentProgram(key: Key): ProgramData | undefined {
  if (!isProgramKey(key) || !key.programs) return undefined
  if (key.current_preset_id === null) {
    return key.programs[0]
  }
  return key.programs.find(p => p.id === key.current_preset_id)
}

function getCurrentSong(key: Key): SongData | undefined {
  if (!isLyricsKey(key) || !key.songs) return undefined
  if (key.current_preset_id === null) {
    return key.songs[0]
  }
  return key.songs.find(s => s.id === key.current_preset_id)
}

function getCurrentLyricText(key: Key): string {
  const song = getCurrentSong(key)
  if (!song || !song.lyrics || song.current_lyric_id === null) {
    return ''
  }
  
  const lyric = song.lyrics.find(l => l.id === song.current_lyric_id)
  return lyric ? lyric.text : ''
}

function shouldShowLyrics(key: Key): boolean {
  if (!isLyricsKey(key)) return false
  
  const song = getCurrentSong(key)
  if (!song || !song.lyrics || song.current_lyric_id === null) {
    return false
  }
  
  return key.status === 'OPENED' || key.status === 'OPENING' || key.status === 'PLAYING_FORWARD'
}

async function animateLyrics(key: Key) {
  if (key.status !== 'PLAYING_FORWARD') return
  
  const song = getCurrentSong(key)
  if (!song || !song.lyrics || song.current_lyric_id === null) return
  
  const lyric = song.lyrics.find(l => l.id === song.current_lyric_id)
  if (!lyric) return
  
  const text = lyric.text
  if (!text || text.trim() === '') return
  
  const transitionTime = Number(lyric.transition_time) * 1000
  const timePerChar = Math.min(transitionTime, 4000) / text.length
  
  // Reset styles
  if (!charStyles.value[key.id]) {
    charStyles.value[key.id] = {}
  }
  charStyles.value[key.id] = {}
  
  await nextTick()
  
  // Animate each character
  for (let i = 0; i < text.length; i++) {
    charStyles.value[key.id][i] = {
      fontSize: '0em',
      opacity: 0,
      transition: `all ${timePerChar * 3}ms ease`
    }
    
    await new Promise(resolve => setTimeout(resolve, 10))
    
    charStyles.value[key.id][i] = {
      fontSize: '1em',
      opacity: 1,
      transition: `all ${timePerChar * 3}ms ease`
    }
    
    await new Promise(resolve => setTimeout(resolve, timePerChar))
  }
}

// Watch for PLAYING_FORWARD status changes
watch(() => keys.value.map(k => ({ id: k.id, status: k.status })), (newVals, oldVals) => {
  newVals.forEach((newVal) => {
    const key = keys.value.find(k => k.id === newVal.id)
    if (key && isLyricsKey(key) && newVal.status === 'PLAYING_FORWARD') {
      const oldVal = oldVals?.find(o => o.id === newVal.id)
      if (!oldVal || oldVal.status !== 'PLAYING_FORWARD') {
        animateLyrics(key)
      }
    }
  })
}, { deep: true })

onMounted(() => {
  connectWebSocket()
})

onUnmounted(() => {
  if (ws) {
    ws.close()
  }
})
</script>

<style scoped>
/* Ensure full screen and no scrolling */
body, html {
  overflow: hidden;
  width: 100vw;
  height: 100vh;
}
</style>
