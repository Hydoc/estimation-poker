<script setup lang="ts">
import { type SendableWebsocketMessageType, useWebsocketStore } from "@/stores/websocket";
import { useRouter } from "vue-router";
import RoomDetail from "@/components/RoomDetail.vue";
import { computed, onMounted } from "vue";

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
const didSkip = computed(() => websocketStore.didSkip);
const showAllGuesses = computed(() => websocketStore.showAllGuesses);
const possibleGuesses = computed(() => websocketStore.possibleGuesses);
const permissions = computed(() => websocketStore.permissions);
const roomIsLocked = computed(() => websocketStore.roomIsLocked);

function sendMessage(
  type: SendableWebsocketMessageType,
  data: string | number | null | { password?: string; key: string },
) {
  websocketStore.send({ type, data });
}

function leaveRoom() {
  websocketStore.disconnect();
  router.push("/");
}

onMounted(async () => {
  await Promise.all([
    websocketStore.fetchPossibleGuesses(),
    websocketStore.fetchPermissions(),
    websocketStore.fetchRoomIsLocked(),
  ]);
});
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
    :did-skip="didSkip"
    :show-all-guesses="showAllGuesses"
    :possible-guesses="possibleGuesses"
    :permissions="permissions"
    :room-is-locked="roomIsLocked"
    @estimate="sendMessage('estimate', $event)"
    @guess="sendMessage('guess', $event)"
    @reveal="sendMessage('reveal', null)"
    @new-round="sendMessage('new-round', null)"
    @leave="leaveRoom"
    @skip="sendMessage('skip', null)"
    @lock-room="sendMessage('lock-room', $event)"
    @open-room="sendMessage('open-room', $event)"
  />
</template>

<style scoped></style>
