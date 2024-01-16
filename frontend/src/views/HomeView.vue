<script setup lang="ts">
import TheRoomChoice from "@/components/TheRoomChoice.vue";
import { useWebsocketStore } from "@/stores/websocket";
const websocketStore = useWebsocketStore();
websocketStore.disconnect();
</script>

<template>
  <main>
    <TheRoomChoice />
  </main>
</template>
