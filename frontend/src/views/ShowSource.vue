<template>
  <div class="fixed inset-0 overflow-hidden bg-transparent">
    <!-- KEY0 - Title Display -->
    <div
      v-show="presets[0] && (presets[0].status === 'OPENED' || presets[0].status === 'OPENING')"
      class="absolute inset-0 flex items-center justify-center"
      :style="{
        transition: presets[0]?.status === 'OPENING' || presets[0]?.status === 'CLOSING'
          ? `opacity ${presets[0]?.transition_time}s ease`
          : 'none',
        opacity: presets[0]?.status === 'OPENED' || presets[0]?.status === 'OPENING' ? 1 : 0
      }"
    >
      <div class="text-center px-8">
        <div class="text-6xl font-bold text-white mb-4" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
          {{ getCurrentContent(presets[0])?.name }}
        </div>
        <div class="text-4xl text-white" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
          {{ getCurrentContent(presets[0])?.person }}
        </div>
      </div>
    </div>

    <!-- KEY1 - Title Display -->
    <div
      v-show="presets[1] && (presets[1].status === 'OPENED' || presets[1].status === 'OPENING')"
      class="absolute inset-0 flex items-center justify-center"
      :style="{
        transition: presets[1]?.status === 'OPENING' || presets[1]?.status === 'CLOSING'
          ? `opacity ${presets[1]?.transition_time}s ease`
          : 'none',
        opacity: presets[1]?.status === 'OPENED' || presets[1]?.status === 'OPENING' ? 1 : 0
      }"
    >
      <div class="text-center px-8">
        <div class="text-6xl font-bold text-white mb-4" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
          {{ getCurrentContent(presets[1])?.name }}
        </div>
        <div class="text-4xl text-white" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
          {{ getCurrentContent(presets[1])?.person }}
        </div>
      </div>
    </div>

    <!-- KEY2 - Lyrics Display -->
    <div
      v-if="presets[2]"
      class="absolute bottom-20 left-0 right-0 flex justify-center"
      :style="{
        transition: shouldShowLyrics(2)
          ? 'none'
          : `opacity ${presets[2]?.transition_time}s ease`,
        opacity: shouldShowLyrics(2) ? 1 : 0
      }"
    >
      <div class="text-5xl font-bold text-white px-8 text-center" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
        <span
          v-for="(char, index) in getCurrentLyricText(2)"
          :key="index"
          :ref="el => { if (el) charRefs[2][index] = el as HTMLElement }"
          class="inline-block"
          :style="charStyles[2]?.[index] || {}"
        >
          {{ char }}
        </span>
      </div>
    </div>

    <!-- KEY3 - Lyrics Display -->
    <div
      v-if="presets[3]"
      class="absolute bottom-20 left-0 right-0 flex justify-center"
      :style="{
        transition: shouldShowLyrics(3)
          ? 'none'
          : `opacity ${presets[3]?.transition_time}s ease`,
        opacity: shouldShowLyrics(3) ? 1 : 0
      }"
    >
      <div class="text-5xl font-bold text-white px-8 text-center" style="text-shadow: 2px 2px 4px rgba(0,0,0,0.8)">
        <span
          v-for="(char, index) in getCurrentLyricText(3)"
          :key="index"
          :ref="el => { if (el) charRefs[3][index] = el as HTMLElement }"
          class="inline-block"
          :style="charStyles[3]?.[index] || {}"
        >
          {{ char }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { createWebSocket } from '@/api'
import type { Preset, PresetContent } from '@/types'

const presets = ref<Preset[]>([])
const reconnectAttempts = ref(0)
const reconnectDelay = 1000
let ws: WebSocket | null = null

const charRefs = ref<Record<number, Record<number, HTMLElement>>>({
  2: {},
  3: {}
})
const charStyles = ref<Record<number, Record<number, any>>>({
  2: {},
  3: {}
})

function connectWebSocket() {
  console.log(`Attempting WebSocket connection (attempt ${reconnectAttempts.value + 1})...`)
  
  try {
    ws = createWebSocket((data: Preset[]) => {
      presets.value = data
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

function getCurrentContent(preset: Preset | undefined): PresetContent | undefined {
  if (!preset) return undefined
  return preset.content[Number(preset.current_preset)]
}

function getCurrentLyricText(keyIndex: number): string {
  const content = getCurrentContent(presets.value[keyIndex])
  if (!content || !content.lyrics || content.current_lyrics === undefined) {
    return ''
  }
  
  const lyric = content.lyrics[content.current_lyrics]
  return lyric ? lyric.text : ''
}

function shouldShowLyrics(keyIndex: number): boolean {
  const preset = presets.value[keyIndex]
  if (!preset) return false
  
  const content = getCurrentContent(preset)
  if (!content || !content.lyrics || content.current_lyrics === undefined) {
    return false
  }
  
  return preset.status === 'OPENED' || preset.status === 'OPENING' || preset.status === 'PLAYING_FORWARD'
}

async function animateLyrics(keyIndex: number) {
  const preset = presets.value[keyIndex]
  if (!preset || preset.status !== 'PLAYING_FORWARD') return
  
  const content = getCurrentContent(preset)
  if (!content || !content.lyrics || content.current_lyrics === undefined) return
  
  const lyric = content.lyrics[content.current_lyrics]
  const text = lyric.text
  if (!text || text.trim() === '') return
  
  const transitionTime = Number(lyric.transition_time) * 1000
  const timePerChar = Math.min(transitionTime, 4000) / text.length
  
  // Reset styles
  charStyles.value[keyIndex] = {}
  
  await nextTick()
  
  // Animate each character
  for (let i = 0; i < text.length; i++) {
    charStyles.value[keyIndex][i] = {
      fontSize: '0em',
      opacity: 0,
      transition: `all ${timePerChar * 3}ms ease`
    }
    
    await new Promise(resolve => setTimeout(resolve, 10))
    
    charStyles.value[keyIndex][i] = {
      fontSize: '1em',
      opacity: 1,
      transition: `all ${timePerChar * 3}ms ease`
    }
    
    await new Promise(resolve => setTimeout(resolve, timePerChar))
  }
}

// Watch for PLAYING_FORWARD status changes
watch(() => presets.value.map(p => ({ status: p.status, keyNum: p.key_num })), (newVals, oldVals) => {
  newVals.forEach((newVal, index) => {
    if (index >= 2 && newVal.status === 'PLAYING_FORWARD') {
      const oldVal = oldVals?.[index]
      if (!oldVal || oldVal.status !== 'PLAYING_FORWARD') {
        animateLyrics(index)
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
