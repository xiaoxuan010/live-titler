<template>
  <div class="min-h-screen bg-white p-4">
    <!-- Connection Status Banner -->
    <div
      v-if="!connected"
      class="fixed top-0 left-0 right-0 bg-red-600 text-white text-center py-2 z-50"
    >
      服务器连接已断开，正在重试...
    </div>

    <div class="container mx-auto max-w-6xl" :class="{ 'mt-12': !connected }">
      <h1 class="text-3xl font-bold mb-6">Live Titler - 控制面板</h1>

      <!-- Key Controls -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Button
          v-for="key in keys"
          :key="key.id"
          :variant="
            key.status === 'OPENED' || key.status === 'OPENING'
              ? 'destructive'
              : 'outline'
          "
          @click="toggleKey(key)"
          :disabled="!connected"
          class="h-16"
        >
          {{ key.name }}
        </Button>
      </div>

      <!-- Progress Bars -->
      <div class="space-y-2 mb-6">
        <div
          v-for="key in keys"
          :key="`progress-${key.id}`"
          class="h-2 bg-gray-100 rounded-full overflow-hidden"
        >
          <div
            class="h-full bg-red-600 transition-all"
            :style="{
              width:
                key.status === 'OPENED' || key.status === 'OPENING'
                  ? '100%'
                  : '0%',
              transitionDuration: key.transition_time + 's',
            }"
          ></div>
        </div>
      </div>

      <!-- Settings Section -->
      <div class="mb-6">
        <h2 class="text-2xl font-bold mb-4">设置</h2>
        <div class="flex gap-2">
          <router-link to="/control-panel/keys" class="inline-block">
            <Button variant="outline"> 管理 Keys </Button>
          </router-link>
        </div>
      </div>

      <!-- Key Panels -->
      <div class="space-y-4">
        <div
          v-for="key in keys"
          :key="`panel-${key.id}`"
          class="border border-gray-200 rounded-lg p-4"
        >
          <h3 class="text-xl font-bold mb-4">{{ key.name }}</h3>

          <!-- For Program Keys -->
          <div v-if="isProgramKey(key)" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-1">转场时间 (s)</label>
              <input
                type="number"
                :value="key.transition_time"
                @input="updateKeyTransition(key, $event)"
                :disabled="!connected"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
              />
            </div>

            <div v-if="getCurrentProgram(key)">
              <label class="block text-sm font-medium mb-1">序号</label>
              <input
                type="text"
                :value="getCurrentProgram(key)!.num"
                @input="
                  updateProgramField(
                    getCurrentProgram(key)!,
                    'num',
                    ($event.target as HTMLInputElement).value,
                  )
                "
                :disabled="!connected"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
              />
            </div>

            <div v-if="getCurrentProgram(key)">
              <label class="block text-sm font-medium mb-1">表演者</label>
              <input
                type="text"
                :value="getCurrentProgram(key)!.person"
                @input="
                  updateProgramField(
                    getCurrentProgram(key)!,
                    'person',
                    ($event.target as HTMLInputElement).value,
                  )
                "
                :disabled="!connected"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
              />
            </div>

            <div v-if="getCurrentProgram(key)">
              <label class="block text-sm font-medium mb-1">节目名</label>
              <input
                type="text"
                :value="getCurrentProgram(key)!.name"
                @input="
                  updateProgramField(
                    getCurrentProgram(key)!,
                    'name',
                    ($event.target as HTMLInputElement).value,
                  )
                "
                :disabled="!connected"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
              />
            </div>
          </div>

          <!-- For Lyrics Keys -->
          <div v-else-if="isLyricsKey(key)" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-1">歌曲</label>
              <select
                :value="key.current_preset_id"
                @change="
                  selectSong(
                    key,
                    Number(($event.target as HTMLSelectElement).value),
                  )
                "
                :disabled="!connected"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
              >
                <option
                  v-for="song in key.songs"
                  :key="song.id"
                  :value="song.id"
                >
                  {{ song.name || `歌曲 ${song.position}` }}
                </option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium mb-1"
                >默认动画时间 (s)</label
              >
              <input
                type="number"
                :value="key.transition_time"
                @input="updateKeyTransition(key, $event)"
                :disabled="!connected"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
              />
            </div>

            <div v-if="getCurrentSong(key)">
              <label class="block text-sm font-medium mb-1">歌词列表</label>
              <div
                class="border border-gray-300 rounded-md max-h-60 overflow-y-auto"
              >
                <div
                  v-for="lyric in getCurrentSong(key)!.lyrics"
                  :key="lyric.id"
                  @click="selectLyric(key, lyric.id)"
                  :class="[
                    'px-3 py-2 cursor-pointer hover:bg-gray-100',
                    getCurrentSong(key)!.current_lyric_id === lyric.id
                      ? 'bg-gray-100'
                      : '',
                  ]"
                >
                  {{ lyric.text || "(空行)" }}
                </div>
              </div>
            </div>

            <div class="flex gap-2">
              <Button @click="lyricsBack(key)" :disabled="!connected" size="sm">
                后退
              </Button>
              <Button
                @click="lyricsForward(key)"
                :disabled="!connected"
                size="sm"
              >
                前进
              </Button>
              <Button
                @click="lyricsPlayForward(key)"
                :disabled="!connected"
                size="sm"
              >
                播放前进
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { api, createWebSocket } from "@/api";
import type { Key, PresetStatus, ProgramData, SongData } from "@/types";
import { isProgramKey, isLyricsKey } from "@/types";
import Button from "@/components/ui/Button.vue";

const keys = ref<Key[]>([]);
const connected = ref(false);
const retryCount = ref(0);
const retryDelay = 1000;
let ws: WebSocket | null = null;

async function loadKeys() {
  try {
    const data = await api.getKeys();
    keys.value = data.sort((a, b) => a.position - b.position);
    connected.value = true;
    retryCount.value = 0;
  } catch (error) {
    console.error("Failed to load keys:", error);
    connected.value = false;

    console.log(
      `Retrying in ${retryDelay}ms (attempt ${retryCount.value + 1})...`,
    );
    retryCount.value++;
    setTimeout(loadKeys, retryDelay);
  }
}

function connectWebSocket() {
  ws = createWebSocket((data: Key[]) => {
    keys.value = data.sort((a, b) => a.position - b.position);
    connected.value = true;
  });

  ws.onclose = () => {
    connected.value = false;
    console.log("WebSocket disconnected, reconnecting...");
    setTimeout(connectWebSocket, retryDelay);
  };
}

async function updateKeyStatus(key: Key, status: PresetStatus) {
  if (!connected.value) return;

  try {
    await api.updateKey(key.id, { status, version: key.version });
    key.version++;
  } catch (error) {
    console.error("Failed to update key status:", error);
    connected.value = false;
    setTimeout(loadKeys, retryDelay);
  }
}

async function updateKeyTransition(key: Key, event: Event) {
  const value = Number((event.target as HTMLInputElement).value);
  if (!connected.value) return;

  try {
    await api.updateKey(key.id, {
      transition_time: value,
      version: key.version,
    });
    key.version++;
  } catch (error) {
    console.error("Failed to update transition time:", error);
  }
}

function toggleKey(key: Key) {
  // Simply send the transition state to the server
  // Server will handle the timing and final state transition
  const newStatus = key.status === "CLOSED" ? "OPENING" : "CLOSING";
  updateKeyStatus(key, newStatus);
}

function getCurrentProgram(key: Key): ProgramData | undefined {
  if (!isProgramKey(key) || !key.programs) return undefined;
  if (key.current_preset_id === null) {
    return key.programs[0];
  }
  return key.programs.find((p) => p.id === key.current_preset_id);
}

function getCurrentSong(key: Key): SongData | undefined {
  if (!isLyricsKey(key) || !key.songs) return undefined;
  if (key.current_preset_id === null) {
    return key.songs[0];
  }
  return key.songs.find((s) => s.id === key.current_preset_id);
}

async function updateProgramField(
  program: ProgramData,
  field: keyof ProgramData,
  value: string,
) {
  if (!connected.value) return;

  try {
    await api.updateProgram(program.id, { [field]: value });
  } catch (error) {
    console.error(`Failed to update program ${field}:`, error);
  }
}

async function selectSong(key: Key, songId: number) {
  if (!connected.value) return;

  try {
    await api.updateKey(key.id, {
      current_preset_id: songId,
      version: key.version,
    });
    key.version++;
  } catch (error) {
    console.error("Failed to select song:", error);
  }
}

async function selectLyric(key: Key, lyricId: number) {
  const song = getCurrentSong(key);
  if (!song || !connected.value) return;

  try {
    await api.updateSong(song.id, { current_lyric_id: lyricId });
  } catch (error) {
    console.error("Failed to select lyric:", error);
  }
}

async function lyricsBack(key: Key) {
  const song = getCurrentSong(key);
  if (!song || !song.lyrics || !connected.value) return;

  const currentIndex = song.lyrics.findIndex(
    (l) => l.id === song.current_lyric_id,
  );
  if (currentIndex > 0) {
    await selectLyric(key, song.lyrics[currentIndex - 1].id);
  }
}

async function lyricsForward(key: Key) {
  const song = getCurrentSong(key);
  if (!song || !song.lyrics || !connected.value) return;

  const currentIndex = song.lyrics.findIndex(
    (l) => l.id === song.current_lyric_id,
  );
  if (currentIndex >= 0 && currentIndex < song.lyrics.length - 1) {
    await selectLyric(key, song.lyrics[currentIndex + 1].id);
  }
}

async function lyricsPlayForward(key: Key) {
  // Start playing forward - server will handle auto-advancing lyrics
  key.status = "PLAYING_FORWARD";
  await updateKeyStatus(key, "PLAYING_FORWARD");
}

onMounted(() => {
  loadKeys();
  connectWebSocket();
});

onUnmounted(() => {
  if (ws) {
    ws.close();
  }
});
</script>
