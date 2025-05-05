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
  <v-app>
    <header>
      <v-app-bar :elevation="2">
        <v-app-bar-title>Estimation Poker</v-app-bar-title>
      </v-app-bar>
    </header>

    <RouterView class="mt-12" />
  </v-app>
</template>
