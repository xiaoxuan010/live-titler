<template>
  <mdui-icon-label-off--outlined
    :slot="props.slot"
    v-if="iconName == 'label-off'"
  />
  <mdui-icon-label :slot="props.slot" v-if="iconName == 'label'" />
  <mdui-icon-subtitles-off--outlined
    :slot="props.slot"
    v-if="iconName == 'subtitles-off'"
  />
  <mdui-icon-subtitles :slot="props.slot" v-if="iconName == 'subtitles'" />
</template>

<script setup lang="ts">
import { isLyricsKey, isProgramKey, Key } from "@/types";
import "@mdui/icons/label-off--outlined.js";
import "@mdui/icons/label.js";
import "@mdui/icons/subtitles-off--outlined.js";
import "@mdui/icons/subtitles.js";
import "mdui/components/button.js";
import { computed } from "vue";

interface Props {
  keyObj: Key;
  slot: string;
}

const props = defineProps<Props>();

const isActive = computed(
  () =>
    props.keyObj.status === "OPENED" ||
    props.keyObj.status === "OPENING" ||
    props.keyObj.status === "PLAYING_FORWARD",
);

const iconName = computed(() => {
  if (isProgramKey(props.keyObj)) {
    return isActive.value ? "label" : "label-off";
  } else if (isLyricsKey(props.keyObj)) {
    return isActive.value ? "subtitles" : "subtitles-off";
  }
});
</script>
