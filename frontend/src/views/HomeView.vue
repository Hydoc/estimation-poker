<script setup lang="ts">
import TheRoomChoice from "@/components/TheRoomChoice.vue";
import { useWebsocketStore } from "@/stores/websocket";
const websocketStore = useWebsocketStore();
websocketStore.disconnect();
websocketStore.resetRound();
</script>

<template>
  <main>
    <the-room-choice />
  </main>
</template>
