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
const roundState = computed(() => websocketStore.roundState);
const ticketToGuess = computed(() => websocketStore.ticketToGuess);
const guess = computed(() => websocketStore.guess);

function estimate(ticket: string) {
  websocketStore.send({
    type: "estimate",
    data: ticket,
  });
}

function doGuess(guess: number) {
  websocketStore.send({
    type: "guess",
    data: guess,
  });
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
    @estimate="estimate"
    @guess="doGuess"
  />
</template>

<style scoped></style>
