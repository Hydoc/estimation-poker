<script setup lang="ts">
import { useWebsocketStore } from "@/stores/websocket";
import { useRouter } from "vue-router";
import { computed } from "vue";
import type { Developer, ProductOwner } from "@/components/types";

const websocketStore = useWebsocketStore();
const router = useRouter();
if (!websocketStore.isConnected) {
  router.push("/");
}
const usersInRoom = computed(() => websocketStore.usersInRoom);
const roomId = computed(() => websocketStore.roomId);

function isThisUser(user: ProductOwner | Developer) {
  return user.name === websocketStore.username;
}

function roleForUser(user: ProductOwner | Developer) {
  return user.role === "developer" ? "Entwickler" : "Product Owner";
}
</script>

<template>
  <h1>Raum: {{ roomId }}</h1>
  <span>* = Du</span>
<ul>
  <li v-for="user in usersInRoom">
    {{ user.name }} ({{ roleForUser(user) }})<span v-if="isThisUser(user)">*</span>
  </li>
</ul>
</template>

<style scoped>

</style>