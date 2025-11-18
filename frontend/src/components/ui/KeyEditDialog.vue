<template>
  <mdui-dialog
    :open="open"
    @closed="$emit('closed')"
    close-on-overlay-click
    headline="编辑 Key 内容"
    :style="{ '--mdui-dialog-max-width': '900px' }"
  >
    <div v-if="keyData">
      <!-- Song/Program Selector Dropdown -->
      <div style="margin-bottom: 16px">
        <mdui-dropdown v-if="keyData.key_type === 'lyrics'">
          <mdui-button slot="trigger" variant="outlined" style="width: 100%">
            当前歌曲: {{ currentSong?.name || "无" }}
            <mdui-icon-arrow-drop-down slot="end-icon"></mdui-icon-arrow-drop-down>
          </mdui-button>
          <mdui-menu>
            <mdui-menu-item
              v-for="song in songsOf(keyData)"
              :key="song.id"
              @click="selectSong(song.id)"
            >
              {{ song.name }}
            </mdui-menu-item>
          </mdui-menu>
        </mdui-dropdown>

        <div v-else style="font-size: 14px; color: var(--mdui-color-on-surface-variant)">
          编辑节目列表
        </div>
      </div>

      <!-- Tabs for Visual vs CSV mode -->
      <mdui-tabs :value="activeTab" @change="onTabChange">
        <mdui-tab value="visual">可视化编辑</mdui-tab>
        <mdui-tab value="csv">CSV 编辑</mdui-tab>

        <!-- Visual editing tab -->
        <mdui-tab-panel value="visual" slot="panel">
          <div style="margin: 16px 0">
            <!-- For Program Keys -->
            <div v-if="keyData.key_type === 'program'">
              <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center">
                <h3 style="margin: 0; font-size: 16px; font-weight: 600">节目列表</h3>
                <mdui-button @click="addProgram" variant="outlined">
                  <mdui-icon-add slot="icon"></mdui-icon-add>
                  添加节目
                </mdui-button>
              </div>

              <draggable
                v-model="programs"
                item-key="id"
                @end="onProgramsReordered"
                handle=".drag-handle"
                v-bind="dragOptions"
              >
                <template #item="{ element: prog, index }">
                  <div
                    style="
                      border: 1px solid var(--mdui-color-outline-variant);
                      border-radius: 4px;
                      margin-bottom: 8px;
                      padding: 8px;
                      background: var(--mdui-color-surface-container-low);
                    "
                  >
                    <div style="display: flex; gap: 8px; align-items: center">
                      <mdui-button-icon class="drag-handle" style="cursor: move">
                        <mdui-icon-drag-indicator></mdui-icon-drag-indicator>
                      </mdui-button-icon>

                      <div style="flex: 1; display: grid; grid-template-columns: 100px 1fr 1fr; gap: 8px">
                        <mdui-text-field
                          :value="prog.num"
                          @input="updateProgram(index, 'num', ($event.target as HTMLInputElement).value)"
                          label="序号"
                          variant="outlined"
                        ></mdui-text-field>
                        <mdui-text-field
                          :value="prog.name"
                          @input="updateProgram(index, 'name', ($event.target as HTMLInputElement).value)"
                          label="节目名"
                          variant="outlined"
                        ></mdui-text-field>
                        <mdui-text-field
                          :value="prog.person"
                          @input="updateProgram(index, 'person', ($event.target as HTMLInputElement).value)"
                          label="表演者"
                          variant="outlined"
                        ></mdui-text-field>
                      </div>

                      <mdui-button-icon @click="deleteProgram(index)">
                        <mdui-icon-delete></mdui-icon-delete>
                      </mdui-button-icon>
                    </div>
                  </div>
                </template>
              </draggable>

              <div v-if="programs.length === 0" style="text-align: center; color: var(--mdui-color-on-surface-variant); padding: 24px">
                暂无节目，点击"添加节目"按钮创建
              </div>
            </div>

            <!-- For Lyrics Keys -->
            <div v-else-if="keyData.key_type === 'lyrics' && currentSong">
              <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center">
                <h3 style="margin: 0; font-size: 16px; font-weight: 600">
                  歌词列表: {{ currentSong.name }}
                </h3>
                <div style="display: flex; gap: 8px">
                  <mdui-button @click="deleteSongConfirm" variant="outlined" style="color: var(--mdui-color-error)">
                    <mdui-icon-delete slot="icon"></mdui-icon-delete>
                    删除歌曲
                  </mdui-button>
                  <mdui-button @click="addLyric" variant="outlined">
                    <mdui-icon-add slot="icon"></mdui-icon-add>
                    添加歌词
                  </mdui-button>
                </div>
              </div>

              <draggable
                v-model="lyrics"
                item-key="id"
                @end="onLyricsReordered"
                handle=".drag-handle"
                v-bind="dragOptions"
              >
                <template #item="{ element: lyric, index }">
                  <div
                    style="
                      border: 1px solid var(--mdui-color-outline-variant);
                      border-radius: 4px;
                      margin-bottom: 8px;
                      padding: 8px;
                      background: var(--mdui-color-surface-container-low);
                    "
                  >
                    <div style="display: flex; gap: 8px; align-items: center">
                      <mdui-button-icon class="drag-handle" style="cursor: move">
                        <mdui-icon-drag-indicator></mdui-icon-drag-indicator>
                      </mdui-button-icon>

                      <div style="flex: 1; display: grid; grid-template-columns: 1fr 150px; gap: 8px">
                        <mdui-text-field
                          :value="lyric.text"
                          @input="updateLyric(index, 'text', ($event.target as HTMLInputElement).value)"
                          label="歌词内容"
                          variant="outlined"
                        ></mdui-text-field>
                        <mdui-text-field
                          :value="lyric.transition_time"
                          @input="updateLyric(index, 'transition_time', Number(($event.target as HTMLInputElement).value))"
                          type="number"
                          label="动画时间(s)"
                          variant="outlined"
                          step="0.1"
                        ></mdui-text-field>
                      </div>

                      <mdui-button-icon @click="deleteLyric(index)">
                        <mdui-icon-delete></mdui-icon-delete>
                      </mdui-button-icon>
                    </div>
                  </div>
                </template>
              </draggable>

              <div v-if="lyrics.length === 0" style="text-align: center; color: var(--mdui-color-on-surface-variant); padding: 24px">
                暂无歌词，点击"添加歌词"按钮创建
              </div>
            </div>

            <!-- Add Song button for lyrics keys -->
            <div v-if="keyData.key_type === 'lyrics'" style="margin-top: 16px">
              <mdui-button @click="addSong" variant="tonal" style="width: 100%">
                <mdui-icon-add slot="icon"></mdui-icon-add>
                添加新歌曲
              </mdui-button>
            </div>
          </div>
        </mdui-tab-panel>

        <!-- CSV editing tab -->
        <mdui-tab-panel value="csv" slot="panel">
          <div style="margin: 16px 0">
            <div style="margin-bottom: 12px">
              <strong>CSV 格式说明：</strong>
              <div v-if="keyData.key_type === 'program'" style="font-size: 13px; color: var(--mdui-color-on-surface-variant)">
                每行格式：序号,节目名,表演者
              </div>
              <div v-else style="font-size: 13px; color: var(--mdui-color-on-surface-variant)">
                每行格式：歌词内容,动画时间(秒)
              </div>
            </div>

            <mdui-text-field
              v-model="csvContent"
              label="CSV 内容"
              variant="outlined"
              rows="15"
              style="width: 100%"
            ></mdui-text-field>

            <div style="margin-top: 12px; display: flex; gap: 8px">
              <mdui-button @click="importFromCSV" variant="filled">导入 CSV</mdui-button>
              <mdui-button @click="exportToCSV" variant="outlined">导出 CSV</mdui-button>
            </div>
          </div>
        </mdui-tab-panel>
      </mdui-tabs>
    </div>

    <mdui-button slot="action" variant="text" @click="$emit('closed')">关闭</mdui-button>
    <mdui-button slot="action" variant="filled" @click="saveChanges" :loading="saving">
      保存
    </mdui-button>
  </mdui-dialog>

  <!-- Confirm delete song dialog -->
  <mdui-dialog
    :open="!!deletingSong"
    @closed="deletingSong = null"
    headline="确认删除歌曲"
  >
    <p>确定要删除歌曲 "{{ currentSong?.name }}" 吗？此操作将删除该歌曲的所有歌词，且无法撤销。</p>
    <mdui-button slot="action" variant="text" @click="deletingSong = null">取消</mdui-button>
    <mdui-button slot="action" variant="text" @click="deleteSongExecute">删除</mdui-button>
  </mdui-dialog>

  <!-- Add song dialog -->
  <mdui-dialog
    :open="addingSong"
    @closed="addingSong = false"
    headline="添加新歌曲"
  >
    <mdui-text-field
      v-model="newSongName"
      label="歌曲名称"
      style="width: 100%"
    ></mdui-text-field>
    <mdui-button slot="action" variant="text" @click="addingSong = false">取消</mdui-button>
    <mdui-button slot="action" variant="filled" @click="addSongExecute" :disabled="!newSongName">
      添加
    </mdui-button>
  </mdui-dialog>

  <!-- Snackbar -->
  <mdui-snackbar :open="!!message" @closed="message = ''" closeable>
    {{ message }}
  </mdui-snackbar>
</template>

<script setup lang="ts">
import { api } from "@/api";
import type { Key, LyricData, ProgramData, SongData } from "@/types";
import { songsOf } from "@/types";
import "@mdui/icons/add.js";
import "@mdui/icons/arrow-drop-down.js";
import "@mdui/icons/delete.js";
import "@mdui/icons/drag-indicator.js";
import "mdui/components/button-icon.js";
import "mdui/components/button.js";
import "mdui/components/dialog.js";
import "mdui/components/dropdown.js";
import "mdui/components/menu-item.js";
import "mdui/components/menu.js";
import "mdui/components/snackbar.js";
import "mdui/components/tab-panel.js";
import "mdui/components/tab.js";
import "mdui/components/tabs.js";
import "mdui/components/text-field.js";
import { computed, ref, watch } from "vue";
import draggable from "vuedraggable";

const props = defineProps<{
  open: boolean;
  keyData: Key | null;
}>();

const emit = defineEmits<{
  (e: "closed"): void;
  (e: "updated"): void;
}>();

const activeTab = ref("visual");
const csvContent = ref("");
const saving = ref(false);
const message = ref("");

const currentSongId = ref<number | null>(null);
const programs = ref<ProgramData[]>([]);
const lyrics = ref<LyricData[]>([]);

const deletingSong = ref<number | null>(null);
const addingSong = ref(false);
const newSongName = ref("");

const dragOptions = {
  animation: 200,
};

const currentSong = computed(() => {
  if (!props.keyData || props.keyData.key_type !== "lyrics") return null;
  const songs = songsOf(props.keyData);
  if (currentSongId.value !== null) {
    return songs.find((s) => s.id === currentSongId.value) || songs[0];
  }
  return songs[0];
});

// Watch for keyData changes and initialize
watch(
  () => props.keyData,
  (newKeyData) => {
    if (!newKeyData) return;

    if (newKeyData.key_type === "program") {
      programs.value = [...(newKeyData.programs || [])];
    } else if (newKeyData.key_type === "lyrics") {
      const songs = songsOf(newKeyData);
      if (songs.length > 0) {
        currentSongId.value = songs[0].id;
        lyrics.value = [...(songs[0].lyrics || [])];
      }
    }
  },
  { immediate: true }
);

// Watch for song changes
watch(currentSongId, (newSongId) => {
  if (!props.keyData || props.keyData.key_type !== "lyrics" || !newSongId) return;
  const song = songsOf(props.keyData).find((s) => s.id === newSongId);
  if (song) {
    lyrics.value = [...(song.lyrics || [])];
  }
});

function onTabChange(e: Event) {
  const ev = e as CustomEvent;
  const val = ev?.detail?.value ?? (e.target as any)?.value;
  if (val) {
    activeTab.value = String(val);
  }
}

function selectSong(songId: number) {
  currentSongId.value = songId;
}

// Program operations
function addProgram() {
  programs.value.push({
    id: 0, // Temporary ID
    num: "",
    name: "",
    person: "",
    position: programs.value.length,
  });
}

function updateProgram(index: number, field: keyof ProgramData, value: string) {
  programs.value[index] = { ...programs.value[index], [field]: value };
}

function deleteProgram(index: number) {
  programs.value.splice(index, 1);
}

async function onProgramsReordered() {
  // Update positions
  programs.value.forEach((prog, idx) => {
    prog.position = idx;
  });
}

// Lyric operations
function addLyric() {
  lyrics.value.push({
    id: 0, // Temporary ID
    text: "",
    transition_time: 1.0,
    position: lyrics.value.length,
  });
}

function updateLyric(index: number, field: keyof LyricData, value: string | number) {
  lyrics.value[index] = { ...lyrics.value[index], [field]: value };
}

function deleteLyric(index: number) {
  lyrics.value.splice(index, 1);
}

async function onLyricsReordered() {
  // Update positions
  lyrics.value.forEach((lyric, idx) => {
    lyric.position = idx;
  });
}

// Song operations
function addSong() {
  newSongName.value = "";
  addingSong.value = true;
}

async function addSongExecute() {
  if (!props.keyData || !newSongName.value) return;

  try {
    await api.createSong(props.keyData.id, newSongName.value);
    message.value = "歌曲已添加";
    addingSong.value = false;
    emit("updated");
  } catch (error) {
    message.value = `添加歌曲失败: ${error}`;
  }
}

function deleteSongConfirm() {
  if (currentSong.value) {
    deletingSong.value = currentSong.value.id;
  }
}

async function deleteSongExecute() {
  if (!deletingSong.value) return;

  try {
    await api.deleteSong(deletingSong.value);
    message.value = "歌曲已删除";
    deletingSong.value = null;
    emit("updated");
  } catch (error) {
    message.value = `删除歌曲失败: ${error}`;
  }
}

// CSV operations
async function exportToCSV() {
  if (!props.keyData) return;

  try {
    if (props.keyData.key_type === "program") {
      csvContent.value = await api.exportProgramsCSV(props.keyData.id);
    } else if (props.keyData.key_type === "lyrics" && currentSong.value) {
      csvContent.value = await api.exportLyricsCSV(currentSong.value.id);
    }
    message.value = "已导出到 CSV";
  } catch (error) {
    message.value = `导出失败: ${error}`;
  }
}

async function importFromCSV() {
  if (!props.keyData || !csvContent.value) return;

  saving.value = true;
  try {
    if (props.keyData.key_type === "program") {
      await api.importProgramsCSV(props.keyData.id, csvContent.value);
    } else if (props.keyData.key_type === "lyrics" && currentSong.value) {
      await api.importLyricsCSV(currentSong.value.id, csvContent.value);
    }
    message.value = "CSV 导入成功";
    emit("updated");
  } catch (error) {
    message.value = `导入失败: ${error}`;
  } finally {
    saving.value = false;
  }
}

// Save changes
async function saveChanges() {
  if (!props.keyData) return;

  saving.value = true;
  try {
    if (props.keyData.key_type === "program") {
      // Delete old programs and create new ones
      const existingPrograms = props.keyData.programs || [];
      for (const prog of existingPrograms) {
        await api.deleteProgram(prog.id);
      }

      for (const prog of programs.value) {
        await api.createProgram(props.keyData.id, {
          num: prog.num,
          name: prog.name,
          person: prog.person,
        });
      }
    } else if (props.keyData.key_type === "lyrics" && currentSong.value) {
      // Delete old lyrics and create new ones
      const song = songsOf(props.keyData).find((s) => s.id === currentSong.value!.id);
      if (song) {
        for (const lyric of song.lyrics) {
          await api.deleteLyric(lyric.id);
        }

        for (const lyric of lyrics.value) {
          await api.createLyric(currentSong.value.id, lyric.text, lyric.transition_time);
        }
      }
    }

    message.value = "保存成功";
    emit("updated");
    emit("closed");
  } catch (error) {
    message.value = `保存失败: ${error}`;
  } finally {
    saving.value = false;
  }
}
</script>
