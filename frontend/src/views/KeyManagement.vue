<template>
  <!-- 顶部 App Bar -->
  <mdui-layout full-height>
    <mdui-top-app-bar>
      <mdui-button-icon @click="$router.push('/control-panel')">
        <mdui-icon-arrow-back></mdui-icon-arrow-back>
      </mdui-button-icon>
      <mdui-top-app-bar-title>Key 管理</mdui-top-app-bar-title>
    </mdui-top-app-bar>

    <mdui-layout-main>
      <!-- 新增 Key -->
      <div style="padding: 16px; max-width: 1200px; margin: 0 auto">
        <mdui-card style="padding: 16px; margin-bottom: 16px">
          <h2 style="margin: 0 0 16px 0; font-size: 18px; font-weight: 600">
            创建新 Key
          </h2>
          <div style="display: flex; gap: 12px; align-items: flex-end">
            <mdui-text-field
              v-model="newKeyName"
              label="Key 名称"
              style="flex: 1"
            ></mdui-text-field>

            <mdui-select v-model="newKeyType" style="width: 180px">
              <mdui-menu-item value="program">节目信息</mdui-menu-item>
              <mdui-menu-item value="lyrics">歌词</mdui-menu-item>
            </mdui-select>

            <mdui-button
              @click="createKey"
              :disabled="!newKeyName || loading"
              :loading="loading"
            >
              创建
            </mdui-button>
          </div>
        </mdui-card>

        <!-- Key 列表 -->
        <draggable
          v-model="keys"
          item-key="id"
          @start="drag = true"
          @end="onDragEnd"
          v-bind="dragOptions"
        >
          <template #item="{ element: key }">
            <div
              class="key-item"
              style="
                border: 1px solid var(--mdui-color-outline-variant);
                border-radius: 8px;
                margin-bottom: 8px;
                padding: 8px;
                background: var(--mdui-color-surface);
              "
            >
              <div style="width: 100%">
                <div
                  style="
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 8px;
                  "
                >
                  <div style="display: flex; align-items: center; flex: 1">
                    <!-- Drag handle -->
                    <mdui-button-icon class="drag-handle" style="cursor: move">
                      <mdui-icon-drag-indicator></mdui-icon-drag-indicator>
                    </mdui-button-icon>

                    <div style="flex: 1; margin-left: 8px">
                      <div
                        style="
                          font-size: 16px;
                          font-weight: 600;
                          margin-bottom: 4px;
                        "
                      >
                        {{ key.name }}
                      </div>
                      <div
                        style="
                          font-size: 13px;
                          color: var(--mdui-color-on-surface-variant);
                        "
                      >
                        类型:
                        {{ key.key_type === "program" ? "节目信息" : "歌词" }}
                        | 位置: {{ key.position }} | 状态: {{ key.status }} |
                        版本:
                        {{ key.version }}
                      </div>
                    </div>
                  </div>

                  <div style="display: flex; gap: 8px">
                    <mdui-button-icon
                      @click="startEdit(key)"
                      :disabled="loading"
                    >
                      <mdui-icon-edit></mdui-icon-edit>
                    </mdui-button-icon>

                    <mdui-button-icon
                      @click="confirmDelete(key)"
                      :disabled="loading"
                    >
                      <mdui-icon-delete></mdui-icon-delete>
                    </mdui-button-icon>
                  </div>
                </div>

                <!-- Key details -->
                <div
                  style="
                    font-size: 13px;
                    color: var(--mdui-color-on-surface-variant);
                    margin-top: 8px;
                    margin-left: 48px;
                  "
                >
                  <div
                    v-if="key.key_type === 'program'"
                    style="display: flex; gap: 8px; flex-wrap: wrap"
                  >
                    <strong>节目列表 ({{ programsOf(key).length }}):</strong>
                    <mdui-chip v-for="prog in programsOf(key)" :key="prog.id">{{
                      prog.name
                    }}</mdui-chip>
                  </div>
                  <div
                    v-if="key.key_type === 'lyrics'"
                    style="display: flex; gap: 8px; flex-wrap: wrap"
                  >
                    <strong>歌曲列表 ({{ songsOf(key).length }}):</strong>
                    <mdui-chip v-for="song in songsOf(key)" :key="song.id">{{
                      song.name
                    }}</mdui-chip>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </draggable>
      </div>
    </mdui-layout-main>
  </mdui-layout>

  <!-- 编辑 Key 对话框 -->
  <mdui-dialog
    :open="!!editingKey"
    @closed="cancelEdit"
    close-on-overlay-click
    headline="编辑 Key"
  >
    <mdui-text-field
      v-model="editForm.name"
      label="名称"
      style="margin-bottom: 16px"
    ></mdui-text-field>

    <mdui-select
      v-model="editForm.key_type"
      label="类型"
      style="margin-bottom: 16px"
    >
      <mdui-menu-item value="program">节目信息</mdui-menu-item>
      <mdui-menu-item value="lyrics">歌词</mdui-menu-item>
    </mdui-select>

    <mdui-text-field
      v-model.number="editForm.transition_time"
      type="number"
      label="转场时间 (秒)"
      step="0.1"
      min="0"
    ></mdui-text-field>

    <mdui-button slot="action" variant="text" @click="cancelEdit"
      >取消</mdui-button
    >
    <mdui-button
      slot="action"
      variant="text"
      @click="saveEdit"
      :disabled="!editForm.name"
      >保存</mdui-button
    >
  </mdui-dialog>

  <!-- 删除 Key 确认对话框 -->
  <mdui-dialog
    :open="!!deletingKey"
    @closed="cancelDelete"
    close-on-overlay-click
    headline="确认删除"
  >
    <p v-if="deletingKey">
      确定要删除 Key "{{ deletingKey.name }}"
      吗？此操作将同时删除相关的所有数据，且无法撤销。
    </p>

    <mdui-button slot="action" variant="text" @click="cancelDelete"
      >取消</mdui-button
    >
    <mdui-button slot="action" variant="text" @click="executeDelete"
      >删除</mdui-button
    >
  </mdui-dialog>

  <!-- Snackbar for messages -->
  <mdui-snackbar :open="!!error" placement="top" closeable @closed="error = ''">
    {{ error }}
  </mdui-snackbar>

  <mdui-snackbar
    :open="!!successMessage"
    placement="top"
    closeable
    @closed="successMessage = ''"
  >
    {{ successMessage }}
  </mdui-snackbar>
</template>

<script setup lang="ts">
import { api } from "@/api";
import type { Key, KeyType } from "@/types";
import { programsOf, songsOf } from "@/types";
import "@mdui/icons/arrow-back.js";
import "@mdui/icons/delete.js";
import "@mdui/icons/drag-indicator.js";
import "@mdui/icons/edit.js";
import "mdui/components/button-icon.js";
import "mdui/components/button.js";
import "mdui/components/card.js";
import "mdui/components/chip.js";
import "mdui/components/dialog.js";
import "mdui/components/layout-main.js";
import "mdui/components/layout.js";
import "mdui/components/menu-item.js";
import "mdui/components/select.js";
import "mdui/components/snackbar.js";
import "mdui/components/text-field.js";
import "mdui/components/top-app-bar-title.js";
import "mdui/components/top-app-bar.js";
import { computed, onMounted, ref } from "vue";
import draggable from "vuedraggable";

const keys = ref<Key[]>([]);
const loading = ref(false);
const error = ref("");
const successMessage = ref("");
const drag = ref(false);

// Drag options
const dragOptions = computed(() => ({
  animation: 200,
  disabled: loading.value,
}));

// Create new key
const newKeyName = ref("");
const newKeyType = ref<KeyType>("program");

// Edit key
const editingKey = ref<Key | null>(null);
const editForm = ref({
  name: "",
  key_type: "program" as KeyType,
  transition_time: 1.0,
  version: 0,
});

// Delete key
const deletingKey = ref<Key | null>(null);

async function loadKeys() {
  try {
    error.value = "";
    keys.value = await api.getKeys();
  } catch (err) {
    error.value = `加载失败: ${err}`;
    console.error("Failed to load keys:", err);
  }
}

async function createKey() {
  if (!newKeyName.value) return;

  loading.value = true;
  error.value = "";
  successMessage.value = "";

  try {
    await api.createKey(newKeyName.value, newKeyType.value);
    successMessage.value = `Key "${newKeyName.value}" 创建成功`;
    newKeyName.value = "";
    await loadKeys();
  } catch (err) {
    error.value = `创建失败: ${err}`;
    console.error("Failed to create key:", err);
  } finally {
    loading.value = false;
  }
}

function startEdit(key: Key) {
  editingKey.value = key;
  editForm.value = {
    name: key.name,
    key_type: key.key_type,
    transition_time: key.transition_time,
    version: key.version,
  };
}

function cancelEdit() {
  editingKey.value = null;
}

async function saveEdit() {
  if (!editingKey.value) return;

  loading.value = true;
  error.value = "";
  successMessage.value = "";

  try {
    await api.updateKey(editingKey.value.id, editForm.value);
    successMessage.value = `Key "${editForm.value.name}" 更新成功`;
    editingKey.value = null;
    await loadKeys();
  } catch (err) {
    error.value = `更新失败: ${err}`;
    console.error("Failed to update key:", err);
  } finally {
    loading.value = false;
  }
}

function confirmDelete(key: Key) {
  deletingKey.value = key;
}

function cancelDelete() {
  deletingKey.value = null;
}

async function executeDelete() {
  if (!deletingKey.value) return;

  loading.value = true;
  error.value = "";
  successMessage.value = "";

  const keyName = deletingKey.value.name;

  try {
    await api.deleteKey(deletingKey.value.id);
    successMessage.value = `Key "${keyName}" 已删除`;
    deletingKey.value = null;
    await loadKeys();
  } catch (err) {
    error.value = `删除失败: ${err}`;
    console.error("Failed to delete key:", err);
  } finally {
    loading.value = false;
  }
}

async function onDragEnd() {
  drag.value = false;
  loading.value = true;
  error.value = "";

  try {
    const ids = keys.value.map((k) => k.id);
    await api.reorderKeys(ids);
    await loadKeys();
  } catch (err) {
    error.value = `排序失败: ${err}`;
    console.error("Failed to reorder keys:", err);
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadKeys();
});
</script>
