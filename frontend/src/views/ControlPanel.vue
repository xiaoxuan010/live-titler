<template>
  <mdui-layout full-height>
    <mdui-top-app-bar>
      <mdui-button-icon>
        <mdui-icon-festival--rounded />
      </mdui-button-icon>
      <mdui-top-app-bar-title>
        <span class="title">Live Titler</span>
        <span class="title small">控制面板</span>
      </mdui-top-app-bar-title>
      <div style="flex-grow: 1"></div>
      <mdui-tooltip content="管理 Keys">
        <router-link to="/control-panel/keys">
          <mdui-button-icon>
            <mdui-icon-playlist-add-circle--rounded />
          </mdui-button-icon>
        </router-link>
      </mdui-tooltip>
      <mdui-button-icon @click="darkMode.toggle()">
        <mdui-icon-dark-mode--rounded v-if="darkMode.isDark" />
        <mdui-icon-light-mode--rounded v-else />
      </mdui-button-icon>
    </mdui-top-app-bar>

    <!-- 左侧：仅大屏显示的 Key Controls -->
    <mdui-layout-item placement="left" class="key-rail key-controls">
      <div v-for="key in keys" :key="key.id">
        <KeyButton
          :keyObj="key"
          :connected="connected"
          @toggle="toggleKey"
        ></KeyButton>
      </div>
    </mdui-layout-item>

    <!-- 右侧主内容 -->
    <mdui-layout-main>
      <div class="page-wrap mdui-prose">
        <!-- 小屏时显示的 Key Controls -->
        <p>
          <mdui-card class="key-card-sm key-controls grid">
            <div v-for="key in keys" :key="'sm-' + key.id">
              <KeyButton
                :keyObj="key"
                :connected="connected"
                @toggle="toggleKey"
              ></KeyButton>
            </div>
          </mdui-card>
        </p>

        <!-- 设置栏 -->
        <p>
          <mdui-card class="section-card">
            <div style="display: flex; gap: 8px; align-items: center">
              <h2 style="margin: 0; font-size: 20px; font-weight: 600">设置</h2>
              <router-link to="/control-panel/keys">
                <mdui-button variant="outlined">管理 Keys</mdui-button>
              </router-link>
            </div>
          </mdui-card>
        </p>

        <!-- Key Panels with Tabs -->
        <mdui-tabs
          :value="activeTab"
          variant="secondary"
          full-width
          @change="onTabChange"
        >
          <!-- Tab headers -->
          <mdui-tab
            v-for="key in keys"
            :key="`tab-${key.id}`"
            :value="String(key.id)"
          >
            <KeyStatusIcon :key-obj="key" slot="icon" />
            {{ key.name }}
          </mdui-tab>

          <!-- Tab panels -->
          <mdui-tab-panel
            v-for="key in keys"
            :key="`panel-${key.id}`"
            :value="String(key.id)"
            slot="panel"
          >
            <!-- For Program Keys -->
            <div
              v-if="isProgramKey(key)"
              style="
                display: flex;
                flex-direction: column;
                gap: 12px;
                padding: 16px 0;
              "
            >
              <mdui-text-field
                type="number"
                label="转场时间 (s)"
                :value="key.transition_time"
                @input="updateKeyTransition(key, $event)"
                :disabled="!connected"
              ></mdui-text-field>

              <template v-if="getCurrentProgram(key)">
                <mdui-text-field
                  label="序号"
                  :value="getCurrentProgram(key)!.num"
                  @input="
                    updateProgramField(
                      getCurrentProgram(key)!,
                      'num',
                      ($event.target as HTMLInputElement).value,
                    )
                  "
                  :disabled="!connected"
                ></mdui-text-field>

                <mdui-text-field
                  label="表演者"
                  :value="getCurrentProgram(key)!.person"
                  @input="
                    updateProgramField(
                      getCurrentProgram(key)!,
                      'person',
                      ($event.target as HTMLInputElement).value,
                    )
                  "
                  :disabled="!connected"
                ></mdui-text-field>

                <mdui-text-field
                  label="节目名"
                  :value="getCurrentProgram(key)!.name"
                  @input="
                    updateProgramField(
                      getCurrentProgram(key)!,
                      'name',
                      ($event.target as HTMLInputElement).value,
                    )
                  "
                  :disabled="!connected"
                ></mdui-text-field>
              </template>
            </div>

            <!-- For Lyrics Keys -->
            <div
              v-else-if="isLyricsKey(key)"
              style="
                display: flex;
                flex-direction: column;
                gap: 12px;
                padding: 16px 0;
              "
            >
              <mdui-select
                label="当前歌曲"
                :value="
                  key.current_preset_id === null
                    ? undefined
                    : String(key.current_preset_id)
                "
                @change="onSelectSongChange(key, $event)"
                :disabled="!connected"
              >
                <mdui-menu-item
                  v-for="song in key.songs"
                  :key="song.id"
                  :value="String(song.id)"
                >
                  {{ song.name || `歌曲 ${song.position}` }}
                </mdui-menu-item>
              </mdui-select>

              <mdui-text-field
                type="number"
                label="默认动画时间 (s)"
                :value="key.transition_time"
                @input="updateKeyTransition(key, $event)"
                :disabled="!connected"
              ></mdui-text-field>

              <div v-if="getCurrentSong(key)">
                <mdui-list>
                  <mdui-list-subheader>歌词列表</mdui-list-subheader>
                  <mdui-list-item
                    v-for="lyric in getCurrentSong(key)!.lyrics"
                    :key="lyric.id"
                    @click="selectLyric(key, lyric.id)"
                    :active="getCurrentSong(key)!.current_lyric_id === lyric.id"
                  >
                    {{ lyric.text || "(空行)" }}
                    <span slot="end-icon"
                      >转场时间：{{ lyric.transition_time }}s</span
                    >
                  </mdui-list-item>
                </mdui-list>
              </div>

              <div style="display: flex; gap: 8px">
                <mdui-button
                  @click="lyricsBack(key)"
                  :disabled="!connected"
                  variant="outlined"
                  >后退</mdui-button
                >
                <mdui-button
                  @click="lyricsForward(key)"
                  :disabled="!connected"
                  variant="outlined"
                  >前进</mdui-button
                >

                <mdui-button
                  title="通过动画方式播放下一行歌词，仅在 Key 已开启时可用"
                  @click="lyricsPlayForward(key)"
                  :disabled="!connected || key.status != 'OPENED'"
                  variant="filled"
                  >播放前进</mdui-button
                >
              </div>
            </div>
          </mdui-tab-panel>
        </mdui-tabs>
      </div>
    </mdui-layout-main>
  </mdui-layout>

  <!-- 正在连接提示 Snackbar (首次连接失败) -->
  <mdui-snackbar
    :open="retryCount > 0 && !everConnected"
    placement="top"
    auto-close-delay="0"
  >
    正在连接服务器...（{{ retryCount }}）
  </mdui-snackbar>

  <!-- 断联提示 Snackbar (曾经连接成功过) -->
  <mdui-snackbar
    :open="!connected && everConnected"
    placement="top"
    auto-close-delay="0"
  >
    服务器连接已断开，正在重试...（{{ retryCount }}）
  </mdui-snackbar>

  <!-- 连接/重连成功提示 Snackbar -->
  <mdui-snackbar
    :open="reconnectedMessage"
    placement="top"
    closeable
    @closed="reconnectedMessage = false"
    auto-close-delay="3000"
  >
    {{ everConnectedBefore ? "服务器连接已恢复" : "连接成功" }}
  </mdui-snackbar>

  <!-- 409 Conflict Error Snackbar -->
  <mdui-snackbar
    :open="!!conflictError"
    placement="top"
    closeable
    @closed="conflictError = ''"
  >
    {{ conflictError }}
  </mdui-snackbar>
</template>

<script setup lang="ts">
import { api, ApiConflictError, createWebSocket } from "@/api";
import KeyButton from "@/components/ui/KeyButton.vue";
import KeyStatusIcon from "@/components/ui/KeyStatusIcon.vue";
import { useDarkModeStore } from "@/stores/darkMode";
import type { Key, PresetStatus, ProgramData, SongData } from "@/types";
import { isLyricsKey, isProgramKey, programsOf, songsOf } from "@/types";
import "@mdui/icons/dark-mode--rounded.js";
import "@mdui/icons/festival--rounded.js";
import "@mdui/icons/light-mode--rounded.js";
import "@mdui/icons/playlist-add-circle--rounded.js";
import "@mdui/icons/subtitles-off.js";
import "@mdui/icons/subtitles.js";
import "mdui/components/button.js";
import "mdui/components/card.js";
import "mdui/components/layout-item.js";
import "mdui/components/layout-main.js";
import "mdui/components/layout.js";
import "mdui/components/list-item.js";
import "mdui/components/list.js";
import "mdui/components/menu-item.js";
import "mdui/components/select.js";
import "mdui/components/snackbar.js";
import "mdui/components/tab-panel.js";
import "mdui/components/tab.js";
import "mdui/components/tabs.js";
import "mdui/components/text-field.js";
import "mdui/components/tooltip.js";
import "mdui/components/top-app-bar-title.js";
import "mdui/components/top-app-bar.js";
import { onMounted, onUnmounted, ref } from "vue";

const keys = ref<Key[]>([]);
const connected = ref(false);
const everConnected = ref(false); // 是否曾经连接成功过
const everConnectedBefore = ref(false); // 在本次连接成功前是否曾经连接过(用于提示文本)
const retryCount = ref(0);
const reconnectedMessage = ref(false);
const conflictError = ref(""); // 用于409冲突的 snackbar
const activeTab = ref<string>(); // 当前激活的 tab
const retryDelay = 1000;
let ws: WebSocket | null = null;

const darkMode = useDarkModeStore();

async function loadKeys() {
  try {
    const data = await api.getKeys();
    keys.value = data.sort((a, b) => a.position - b.position);

    // 设置默认激活第一个 tab
    if (keys.value.length > 0 && !activeTab.value) {
      activeTab.value = String(keys.value[0].id);
    }

    const wasDisconnected = retryCount.value > 0;
    connected.value = true;
    everConnected.value = true;
    retryCount.value = 0;

    // 如果之前有重试,说明是首次连接成功或重连成功
    if (wasDisconnected) {
      reconnectedMessage.value = true;
      everConnectedBefore.value = true; // 标记为已经历过连接
    }
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

    const wasDisconnected = !connected.value && everConnected.value;
    connected.value = true;
    everConnected.value = true;

    // WebSocket 重连成功
    if (wasDisconnected) {
      reconnectedMessage.value = true;
      retryCount.value = 0;
    }
  });

  ws.onclose = () => {
    connected.value = false;
    retryCount.value++;
    console.log("WebSocket disconnected, reconnecting...");
    setTimeout(connectWebSocket, retryDelay);
  };
}

// Handle tab change emitted by mdui-tabs
function onTabChange(e: Event) {
  const ev = e as CustomEvent;
  const val = ev?.detail?.value ?? (e.target as any)?.value;
  if (val !== undefined && val !== null) {
    activeTab.value = String(val);
  }
}

async function updateKeyStatus(key: Key, status: PresetStatus) {
  if (!connected.value) return;

  try {
    await api.updateKey(key.id, { status, version: key.version });
    key.version++;
  } catch (error) {
    if (error instanceof ApiConflictError) {
      conflictError.value = error.message;
    } else {
      console.error("Failed to update key status:", error);
      connected.value = false;
      setTimeout(loadKeys, retryDelay);
    }
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
    if (error instanceof ApiConflictError) {
      conflictError.value = error.message;
    } else {
      console.error("Failed to update transition time:", error);
    }
  }
}

function toggleKey(key: Key) {
  // Simply send the transition state to the server
  // Server will handle the timing and final state transition
  let newStatus: PresetStatus;
  switch (key.status) {
    case "CLOSED":
      newStatus = "OPENING";
      break;
    case "OPENED":
      newStatus = "CLOSING";
      break;
    case "OPENING":
      newStatus = "CLOSING";
      break;
    case "CLOSING":
      newStatus = "OPENING";
      break;
    case "PLAYING_FORWARD":
      newStatus = "CLOSING";
      break;
  }
  updateKeyStatus(key, newStatus);
}

function getCurrentProgram(key: Key): ProgramData | undefined {
  if (!isProgramKey(key)) return undefined;
  const progs = programsOf(key);
  if (progs.length === 0) return undefined;
  if (key.current_preset_id === null) {
    return progs[0];
  }
  return progs.find((p) => p.id === key.current_preset_id);
}

function getCurrentSong(key: Key): SongData | undefined {
  if (!isLyricsKey(key)) return undefined;
  const songs = songsOf(key);
  if (songs.length === 0) return undefined;
  if (key.current_preset_id === null) {
    return songs[0];
  }
  return songs.find((s) => s.id === key.current_preset_id);
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
    if (error instanceof ApiConflictError) {
      conflictError.value = error.message;
    } else {
      console.error("Failed to select song:", error);
    }
  }
}

function onSelectSongChange(key: Key, e: Event) {
  const raw = (e.target as any).value;
  const val = Array.isArray(raw) ? raw[0] : raw;
  const id = Number(val);
  if (!Number.isNaN(id)) {
    selectSong(key, id);
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

<style scoped>
@font-face {
  font-family: "Roboto";
  src: url("/Roboto-Regular.woff2") format("woff2");
}

mdui-layout {
  font-family:
    Roboto,
    Noto Sans SC,
    PingFang SC,
    Lantinghei SC,
    Microsoft Yahei,
    Hiragino Sans GB,
    "Microsoft Sans Serif",
    WenQuanYi Micro Hei,
    sans-serif;
}

.title.small {
  font-size: 93%;
  margin-left: 1rem;
}

.page-wrap {
  padding: 16px;
  max-width: 1200px;
  margin: 0 auto;
}

/* 左侧 Key 控制区（大屏显示） */
.key-rail {
  width: 280px;
  padding: 16px 8px;
}

.key-controls {
  gap: 12px;
}

.key-controls mdui-button {
  font-size: var(--mdui-typescale-headline-medium-size);
}

.key-card-sm {
  padding: 12px;
}

.key-controls.grid {
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
}

/* 断点示例：>= 1024px 视为大屏，展示左侧栏 */
@media (min-width: 1024px) {
  .key-card-sm {
    display: none;
  }

  .key-rail.key-controls {
    display: flex;
    flex-direction: column;
  }
}

@media (max-width: 1023px) {
  .key-rail {
    display: none;
  }

  /* 小屏时，Key Controls 使用网格布局 */
  .key-controls.grid {
    display: grid;
  }
}

/* 左侧纵向按钮栈样式 */
.section-card {
  padding: 12px;
  margin-bottom: 16px;
}
</style>
