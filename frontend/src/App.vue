<script setup lang="ts">
import { RouterView } from "vue-router";
import { useWebsocketStore } from "@/stores/websocket";
import { onBeforeUnmount } from "vue";
const websocketStore = useWebsocketStore();
onBeforeUnmount(() => {
  websocketStore.disconnect();
});
</script>

<template>
  <header>
    <div class="wrapper">
      <h1>Estimation Poker</h1>
    </div>
  </header>

  <RouterView />
</template>

<style scoped></style>
