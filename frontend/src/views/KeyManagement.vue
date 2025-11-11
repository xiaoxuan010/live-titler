<template>
  <div class="min-h-screen bg-white p-4">
    <!-- Header -->
    <div class="container mx-auto max-w-6xl">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-3xl font-bold">Key 管理</h1>
        <router-link
          to="/control-panel"
          class="px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-md"
        >
          返回控制面板
        </router-link>
      </div>

      <!-- Error message -->
      <div
        v-if="error"
        class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded-md"
      >
        {{ error }}
      </div>

      <!-- Success message -->
      <div
        v-if="successMessage"
        class="mb-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded-md"
      >
        {{ successMessage }}
      </div>

      <!-- Create new key form -->
      <div class="mb-6 p-4 border border-gray-200 rounded-lg">
        <h2 class="text-xl font-bold mb-4">创建新 Key</h2>
        <div class="flex gap-4">
          <input
            v-model="newKeyName"
            type="text"
            placeholder="Key 名称"
            class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-400"
          />
          <select
            v-model="newKeyType"
            class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-400"
          >
            <option value="program">节目信息</option>
            <option value="lyrics">歌词</option>
          </select>
          <Button @click="createKey" :disabled="!newKeyName || loading">
            创建
          </Button>
        </div>
      </div>

      <!-- Keys list -->
      <div class="space-y-4">
        <div
          v-for="(key, index) in keys"
          :key="key.id"
          class="border border-gray-200 rounded-lg p-4"
        >
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-4 flex-1">
              <div>
                <h3 class="text-lg font-bold">{{ key.name }}</h3>
                <p class="text-sm text-gray-600">
                  类型: {{ key.key_type === "program" ? "节目信息" : "歌词" }} |
                  位置: {{ key.position }} | 状态: {{ key.status }} | 版本:
                  {{ key.version }}
                </p>
              </div>
            </div>

            <div class="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                @click="moveUp(index)"
                :disabled="index === 0 || loading"
              >
                ↑ 上移
              </Button>
              <Button
                size="sm"
                variant="outline"
                @click="moveDown(index)"
                :disabled="index === keys.length - 1 || loading"
              >
                ↓ 下移
              </Button>
              <Button
                size="sm"
                variant="outline"
                @click="startEdit(key)"
                :disabled="loading"
              >
                编辑
              </Button>
              <Button
                size="sm"
                variant="destructive"
                @click="confirmDelete(key)"
                :disabled="loading"
              >
                删除
              </Button>
            </div>
          </div>

          <!-- Key details -->
          <div class="text-sm text-gray-600 mt-2">
            <div v-if="key.key_type === 'program' && key.programs">
              <strong>节目列表 ({{ key.programs.length }}):</strong>
              <span v-for="prog in key.programs" :key="prog.id" class="ml-2">
                {{ prog.name }}
              </span>
            </div>
            <div v-if="key.key_type === 'lyrics' && key.songs">
              <strong>歌曲列表 ({{ key.songs.length }}):</strong>
              <span v-for="song in key.songs" :key="song.id" class="ml-2">
                {{ song.name }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Edit Dialog -->
    <div
      v-if="editingKey"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      @click.self="cancelEdit"
    >
      <div class="bg-white p-6 rounded-lg max-w-md w-full">
        <h2 class="text-2xl font-bold mb-4">编辑 Key</h2>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">名称</label>
            <input
              v-model="editForm.name"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-400"
            />
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">类型</label>
            <select
              v-model="editForm.key_type"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-400"
            >
              <option value="program">节目信息</option>
              <option value="lyrics">歌词</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">转场时间 (秒)</label>
            <input
              v-model.number="editForm.transition_time"
              type="number"
              step="0.1"
              min="0"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-400"
            />
          </div>
        </div>

        <div class="flex gap-2 mt-6">
          <Button @click="cancelEdit" variant="outline" class="flex-1"
            >取消</Button
          >
          <Button @click="saveEdit" class="flex-1" :disabled="!editForm.name"
            >保存</Button
          >
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Dialog -->
    <div
      v-if="deletingKey"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      @click.self="cancelDelete"
    >
      <div class="bg-white p-6 rounded-lg max-w-md w-full">
        <h2 class="text-2xl font-bold mb-4">确认删除</h2>
        <p class="mb-6">
          确定要删除 Key "{{ deletingKey.name }}"
          吗？此操作将同时删除相关的所有数据，且无法撤销。
        </p>
        <div class="flex gap-2">
          <Button @click="cancelDelete" variant="outline" class="flex-1"
            >取消</Button
          >
          <Button @click="executeDelete" variant="destructive" class="flex-1"
            >删除</Button
          >
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "@/api";
import type { Key, KeyType } from "@/types";
import Button from "@/components/ui/Button.vue";

const keys = ref<Key[]>([]);
const loading = ref(false);
const error = ref("");
const successMessage = ref("");

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

async function moveUp(index: number) {
  if (index === 0) return;

  loading.value = true;
  error.value = "";

  try {
    const newOrder = [...keys.value];
    [newOrder[index - 1], newOrder[index]] = [
      newOrder[index],
      newOrder[index - 1],
    ];

    const ids = newOrder.map((k) => k.id);
    await api.reorderKeys(ids);
    await loadKeys();
  } catch (err) {
    error.value = `移动失败: ${err}`;
    console.error("Failed to reorder keys:", err);
  } finally {
    loading.value = false;
  }
}

async function moveDown(index: number) {
  if (index === keys.value.length - 1) return;

  loading.value = true;
  error.value = "";

  try {
    const newOrder = [...keys.value];
    [newOrder[index], newOrder[index + 1]] = [
      newOrder[index + 1],
      newOrder[index],
    ];

    const ids = newOrder.map((k) => k.id);
    await api.reorderKeys(ids);
    await loadKeys();
  } catch (err) {
    error.value = `移动失败: ${err}`;
    console.error("Failed to reorder keys:", err);
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadKeys();
});
</script>
