<template>
  <mdui-icon-subtitles-off--outlined
    :slot="props.slot"
    v-if="iconName === 'subtitles-off'"
  />
  <mdui-icon-subtitles :slot="props.slot" v-if="iconName === 'subtitles'" />
</template>

<script setup lang="ts">
import { Key } from "@/types";
import "@mdui/icons/subtitles-off--outlined.js";
import "@mdui/icons/subtitles.js";
import "mdui/components/button.js";
import { computed } from "vue";

interface Props {
  keyObj: Key;
  slot: string;
}

const props = defineProps<Props>();

const iconName = computed(() =>
  props.keyObj.status === "OPENED" || props.keyObj.status === "OPENING"
    ? "subtitles"
    : "subtitles-off",
);
</script>
