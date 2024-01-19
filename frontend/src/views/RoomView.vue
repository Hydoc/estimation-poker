<script setup lang="ts">
import { type SendableWebsocketMessageType, useWebsocketStore} from "@/stores/websocket";
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
const roundState = computed(() => websocketStore.roundState);
const ticketToGuess = computed(() => websocketStore.ticketToGuess);
const guess = computed(() => websocketStore.guess);

function sendMessage(type: SendableWebsocketMessageType, data: string | number | null) {
  websocketStore.send({ type, data });
}
</script>

<template>
  <room-detail
    :current-username="currentUsername"
    :room-id="roomId"
    :users-in-room="usersInRoom"
    :user-role="userRole"
    :round-state="roundState"
    :ticket-to-guess="ticketToGuess"
    :guess="guess"
    @estimate="sendMessage('estimate', $event)"
    @guess="sendMessage('guess', $event)"
    @reveal="sendMessage('reveal', 'blub')"
  />
</template>

<style scoped></style>
