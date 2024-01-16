<script setup lang="ts">
import { useWebsocketStore } from "@/stores/websocket";
import { useRouter } from "vue-router";
import RoomDetail from "@/components/RoomDetail.vue";
import { computed } from "vue";

const websocketStore = useWebsocketStore();
const router = useRouter();
if (!websocketStore.isConnected) {
  router.push("/");
}
const usersInRoom = computed(() => websocketStore.usersInRoom);
const roomId = computed(() => websocketStore.roomId);
const currentUsername = computed(() => websocketStore.username);
const userRole = computed(() => websocketStore.userRole);
</script>

<template>
  <RoomDetail
    :current-username="currentUsername"
    :room-id="roomId"
    :users-in-room="usersInRoom"
    :user-role="userRole"
  />
</template>

<style scoped></style>
