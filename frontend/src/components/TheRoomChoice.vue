<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import type { Ref } from "vue";
import { useWebsocketStore } from "@/stores/websocket";
import { Role } from "@/components/types";
import { useRouter } from "vue-router";
import TheActiveRoomOverview from "@/components/TheActiveRoomOverview.vue";
import RoomForm from "@/components/RoomForm.vue";

const router = useRouter();
const roomId = ref("");
const name = ref("");
const activeRooms: Ref<string[]> = ref([]);
const role: Ref<Role> = ref(Role.Empty);
const websocketStore = useWebsocketStore();
const showUserAlreadyExists = ref(false);
const showRoundIsInProgress = ref(false);
const errorMessage = computed(() => {
  if (showUserAlreadyExists.value) {
    return "Ein Benutzer mit diesem Namen existiert in dem Raum bereits.";
  }

  if (showRoundIsInProgress.value) {
    return "Die Runde in diesem Raum hat bereits begonnen.";
  }

  return "";
});

async function connect() {
  showUserAlreadyExists.value = false;
  showRoundIsInProgress.value = false;
  const roundInRoomInProgress = await websocketStore.isRoundInRoomInProgress(roomId.value);
  if (roundInRoomInProgress) {
    showRoundIsInProgress.value = true;
    return;
  }

  const userAlreadyExistsInRoom = await websocketStore.userExistsInRoom(name.value, roomId.value);
  if (userAlreadyExistsInRoom) {
    showUserAlreadyExists.value = true;
    return;
  }
  websocketStore.connect(name.value, role.value, roomId.value);
  await router.push("/room");
}

async function fetchActiveRooms() {
  activeRooms.value = await websocketStore.fetchActiveRooms();
}

onMounted(fetchActiveRooms);
</script>

<template>
  <v-container>
    <v-row align="center" justify="center">
      <v-col>
        <v-card prepend-icon="mdi-poker-chip">
          <template #title> Ich brauche noch ein paar Informationen bevor es los geht </template>
          <v-card-text>
            <v-container>
              <room-form
                v-model:role="role"
                v-model:name="name"
                v-model:room-id="roomId"
                :error-message="errorMessage"
                @submit="connect"
              />
            </v-container>
            <v-container v-if="activeRooms.length > 0">
              <the-active-room-overview :active-rooms="activeRooms" />
            </v-container>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
