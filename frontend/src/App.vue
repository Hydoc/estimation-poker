<script setup lang="ts">
import { RouterView } from "vue-router";
import { onBeforeUnmount } from "vue";
import { useEstimationStore } from "@/stores/estimation.ts";
const estimationStore = useEstimationStore();
onBeforeUnmount(() => {
  estimationStore.leaveRoom();
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
