<template>
  <mdui-button
    full-width
    style="height: 5rem"
    :variant="buttonVariant"
    :style="{ transitionDuration: keyObj.transition_time + 's' }"
    :disabled="!connected"
    @click="toggleKey(keyObj)"
  >
    <KeyStatusIcon :key-obj="keyObj" slot="icon" />
    {{ keyObj.name }}
  </mdui-button>
</template>

<script setup lang="ts">
import { Key } from "@/types";
import { computed } from "vue";
import KeyStatusIcon from "./KeyStatusIcon.vue";

interface Props {
  keyObj: Key;
  connected: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits(["toggle"]);

const buttonVariant = computed(() =>
  props.keyObj.status === "OPENED" || props.keyObj.status === "OPENING" || props.keyObj.status === "PLAYING_FORWARD"
    ? "filled"
    : "outlined",
);

function toggleKey(keyObj: Key) {
  emit("toggle", keyObj);
}
</script>

<style scoped>
mdui-button {
  transition-property: all;
}
</style>
