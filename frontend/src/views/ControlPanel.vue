<template>
  <div class="min-h-screen bg-background p-4">
    <!-- Connection Status Banner -->
    <div
      v-if="!connected"
      class="fixed top-0 left-0 right-0 bg-destructive text-destructive-foreground text-center py-2 z-50"
    >
      服务器连接已断开，正在重试...
    </div>

    <div class="container mx-auto max-w-6xl" :class="{ 'mt-12': !connected }">
      <h1 class="text-3xl font-bold mb-6">Live Titler - 控制面板</h1>

      <!-- Key Controls -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Button
          v-for="(preset, index) in presets"
          :key="preset.key_num"
          :variant="preset.status === 'OPENED' || preset.status === 'OPENING' ? 'destructive' : 'outline'"
          @click="toggleKey(index)"
          :disabled="!connected"
          class="h-16"
        >
          {{ preset.key_name }}
        </Button>
      </div>

      <!-- Progress Bars -->
      <div class="space-y-2 mb-6">
        <div
          v-for="(preset, index) in presets"
          :key="`progress-${preset.key_num}`"
          class="h-2 bg-secondary rounded-full overflow-hidden"
        >
          <div
            class="h-full bg-destructive transition-all"
            :style="{
              width: preset.status === 'OPENED' || preset.status === 'OPENING' ? '100%' : '0%',
              transitionDuration: preset.transition_time + 's'
            }"
          ></div>
        </div>
      </div>

      <!-- Settings Section -->
      <div class="mb-6">
        <h2 class="text-2xl font-bold mb-4">设置</h2>
        <div class="flex gap-2">
          <Button @click="showImportExport = true" :disabled="!connected">
            导入/导出设置
          </Button>
          <Button variant="outline" @click="clearData" :disabled="!connected">
            清除数据
          </Button>
        </div>
      </div>

      <!-- Key Panels -->
      <div class="space-y-4">
        <div
          v-for="(preset, index) in presets"
          :key="`panel-${preset.key_num}`"
          class="border rounded-lg p-4"
        >
          <h3 class="text-xl font-bold mb-4">{{ preset.key_name }}</h3>
          
          <!-- For KEY0 and KEY1 - Simple presets -->
          <div v-if="index < 2" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-1">转场时间 (s)</label>
              <input
                type="number"
                v-model="preset.transition_time"
                @change="savePreset(index)"
                :disabled="!connected"
                class="w-full px-3 py-2 border rounded-md"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium mb-1">序号</label>
              <input
                type="text"
                v-model="preset.content[Number(preset.current_preset)]!.num"
                @change="savePreset(index)"
                :disabled="!connected"
                class="w-full px-3 py-2 border rounded-md"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium mb-1">表演者</label>
              <input
                type="text"
                v-model="preset.content[Number(preset.current_preset)]!.person"
                @change="savePreset(index)"
                :disabled="!connected"
                class="w-full px-3 py-2 border rounded-md"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium mb-1">节目名</label>
              <input
                type="text"
                v-model="preset.content[Number(preset.current_preset)]!.name"
                @change="savePreset(index)"
                :disabled="!connected"
                class="w-full px-3 py-2 border rounded-md"
              />
            </div>
          </div>

          <!-- For KEY2 and KEY3 - Lyrics -->
          <div v-else class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-1">歌曲</label>
              <select
                v-model="preset.current_preset"
                @change="savePreset(index)"
                :disabled="!connected"
                class="w-full px-3 py-2 border rounded-md"
              >
                <option
                  v-for="(content, cidx) in preset.content"
                  :key="cidx"
                  :value="cidx"
                >
                  {{ content.song_name || `歌曲 ${cidx}` }}
                </option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium mb-1">默认动画时间 (s)</label>
              <input
                type="number"
                v-model="preset.transition_time"
                @change="savePreset(index)"
                :disabled="!connected"
                class="w-full px-3 py-2 border rounded-md"
              />
            </div>

            <div v-if="getCurrentContent(preset)">
              <label class="block text-sm font-medium mb-1">歌词列表</label>
              <div class="border rounded-md max-h-60 overflow-y-auto">
                <div
                  v-for="(lyric, lidx) in getCurrentContent(preset)!.lyrics"
                  :key="lidx"
                  @click="selectLyric(index, lidx)"
                  :class="[
                    'px-3 py-2 cursor-pointer hover:bg-accent',
                    getCurrentContent(preset)!.current_lyrics === lidx ? 'bg-accent' : ''
                  ]"
                >
                  {{ lyric.text || '(空行)' }}
                </div>
              </div>
            </div>

            <div class="flex gap-2">
              <Button @click="lyricsBack(index)" :disabled="!connected" size="sm">
                后退
              </Button>
              <Button @click="lyricsForward(index)" :disabled="!connected" size="sm">
                前进
              </Button>
              <Button @click="lyricsPlayForward(index)" :disabled="!connected" size="sm">
                播放前进
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Import/Export Dialog -->
    <div
      v-if="showImportExport"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      @click.self="showImportExport = false"
    >
      <div class="bg-background p-6 rounded-lg max-w-2xl w-full max-h-[80vh] overflow-y-auto">
        <h2 class="text-2xl font-bold mb-4">配置文件</h2>
        <textarea
          v-model="importExportText"
          class="w-full h-64 px-3 py-2 border rounded-md font-mono text-sm"
          placeholder="在此处输入JSON配置文件"
        ></textarea>
        <div class="flex gap-2 mt-4">
          <Button @click="showImportExport = false" variant="outline">取消</Button>
          <Button @click="copyToClipboard">拷贝到剪贴板</Button>
          <Button @click="importConfig">保存配置并关闭</Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { api, createWebSocket } from '@/api'
import type { Preset, PresetContent } from '@/types'
import Button from '@/components/ui/Button.vue'

const presets = ref<Preset[]>([])
const connected = ref(false)
const retryCount = ref(0)
const retryDelay = 1000
let ws: WebSocket | null = null

const showImportExport = ref(false)
const importExportText = ref('')

async function loadPresets() {
  try {
    const data = await api.getPresets()
    presets.value = data
    connected.value = true
    retryCount.value = 0
  } catch (error) {
    console.error('Failed to load presets:', error)
    connected.value = false
    
    console.log(`Retrying in ${retryDelay}ms (attempt ${retryCount.value + 1})...`)
    retryCount.value++
    setTimeout(loadPresets, retryDelay)
  }
}

function connectWebSocket() {
  ws = createWebSocket((data: Preset[]) => {
    presets.value = data
    connected.value = true
  })
  
  ws.onclose = () => {
    connected.value = false
    console.log('WebSocket disconnected, reconnecting...')
    setTimeout(connectWebSocket, retryDelay)
  }
}

async function savePreset(index: number) {
  if (!connected.value) return
  
  try {
    await api.updatePreset(presets.value[index].key_num, presets.value[index])
  } catch (error) {
    console.error('Failed to save preset:', error)
    connected.value = false
    setTimeout(loadPresets, retryDelay)
  }
}

function toggleKey(index: number) {
  const preset = presets.value[index]
  preset.status = preset.status === 'CLOSED' ? 'OPENING' : 'CLOSING'
  
  savePreset(index)
  
  setTimeout(() => {
    preset.status = preset.status === 'OPENING' ? 'OPENED' : 'CLOSED'
    savePreset(index)
  }, Number(preset.transition_time) * 1000)
}

function getCurrentContent(preset: Preset): PresetContent | undefined {
  return preset.content[Number(preset.current_preset)]
}

function selectLyric(keyIndex: number, lyricIndex: number) {
  const content = getCurrentContent(presets.value[keyIndex])
  if (content && content.lyrics) {
    content.current_lyrics = lyricIndex
    savePreset(keyIndex)
  }
}

function lyricsBack(keyIndex: number) {
  const content = getCurrentContent(presets.value[keyIndex])
  if (content && content.current_lyrics !== undefined && content.current_lyrics > 0) {
    content.current_lyrics--
    savePreset(keyIndex)
  }
}

function lyricsForward(keyIndex: number) {
  const content = getCurrentContent(presets.value[keyIndex])
  if (content && content.lyrics && content.current_lyrics !== undefined) {
    if (content.current_lyrics < content.lyrics.length - 1) {
      content.current_lyrics++
      savePreset(keyIndex)
    }
  }
}

function lyricsPlayForward(keyIndex: number) {
  const preset = presets.value[keyIndex]
  lyricsForward(keyIndex)
  preset.status = 'PLAYING_FORWARD'
  savePreset(keyIndex)
  
  const content = getCurrentContent(preset)
  if (content && content.lyrics && content.current_lyrics !== undefined) {
    const lyric = content.lyrics[content.current_lyrics]
    const time = Number(lyric.transition_time) * 1000
    
    setTimeout(() => {
      preset.status = 'OPENED'
      savePreset(keyIndex)
    }, Math.min(time, 4000))
  }
}

function openImportExport() {
  importExportText.value = JSON.stringify(presets.value, null, 2)
  showImportExport.value = true
}

function copyToClipboard() {
  navigator.clipboard.writeText(importExportText.value)
  alert('已复制到剪贴板')
}

async function importConfig() {
  try {
    const data = JSON.parse(importExportText.value)
    await api.importPresets(data)
    showImportExport.value = false
    await loadPresets()
    alert('导入成功')
  } catch (error) {
    console.error('Failed to import:', error)
    alert('导入失败：' + error)
  }
}

function clearData() {
  if (confirm('确定要清除所有数据吗？')) {
    // This would need backend support to reset to defaults
    alert('此功能需要后端支持')
  }
}

onMounted(() => {
  loadPresets()
  connectWebSocket()
})

onUnmounted(() => {
  if (ws) {
    ws.close()
  }
})
</script>
