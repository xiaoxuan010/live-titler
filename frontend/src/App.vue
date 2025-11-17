<template>
  <div id="app" :class="{ 'mdui-theme-dark': darkMode.isDark }">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { useDarkModeStore } from "@/stores/darkMode";

const darkMode = useDarkModeStore();
</script>

<style>
#app {
  width: 100%;
  height: 100vh;
}
</style>
