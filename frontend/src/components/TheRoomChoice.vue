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
const passwordForRoom = ref("");
const showPasswordDialog = ref(false);
const websocketStore = useWebsocketStore();
const showUserAlreadyExists = ref(false);
const showRoundIsInProgress = ref(false);
const showPasswordDoesNotMatch = ref(false);
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

  const isLocked = await websocketStore.isRoomLocked(roomId.value);
  if (isLocked && passwordForRoom.value === "") {
    showPasswordDialog.value = true;
    return;
  }

  const passwordMatches = isLocked
    ? await websocketStore.passwordMatchesRoom(roomId.value, passwordForRoom.value)
    : true;
  if (isLocked && !passwordMatches) {
    showPasswordDialog.value = true;
    showPasswordDoesNotMatch.value = true;
    return;
  }

  showPasswordDialog.value = false;

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

async function setFieldsAndConnect(roomToJoin: string, passedName: string, passedRole: Role) {
  roomId.value = roomToJoin;
  name.value = passedName;
  role.value = passedRole;
  await connect();
}

async function fetchActiveRooms() {
  activeRooms.value = await websocketStore.fetchActiveRooms();
}

onMounted(fetchActiveRooms);
</script>

<template>
  <v-container>
    <v-dialog width="500" v-model="showPasswordDialog">
      <v-card>
        <v-card-title>Für diesen Raum wird ein Passwort benötigt</v-card-title>
        <v-card-text>
          <v-text-field type="password" placeholder="Passwort" v-model="passwordForRoom" />
          <v-alert
            v-if="showPasswordDoesNotMatch"
            color="error"
            text="Passwort stimmt nicht überein"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="showPasswordDialog = false">Abbrechen</v-btn>
          <v-btn color="green" @click="connect">Ok</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
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
              <the-active-room-overview :active-rooms="activeRooms" @join="setFieldsAndConnect" />
            </v-container>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
