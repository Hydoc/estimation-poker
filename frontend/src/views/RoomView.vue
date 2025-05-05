<script setup lang="ts">
import { type SendableWebsocketMessageType, useWebsocketStore } from "@/stores/websocket";
import { useRouter } from "vue-router";
import RoomDetail from "@/components/RoomDetail.vue";
import { ref, computed, onMounted } from "vue";
import { RoundState } from "@/components/types.ts";

const websocketStore = useWebsocketStore();
const router = useRouter();
if (!websocketStore.isConnected) {
  router.push("/");
}
const showSetRoomPasswordDialog = ref(false);
const showPassword = ref(false);
const showSnackbar = ref(false);
const roomPassword = ref("");
const snackbarText = ref("");
const usersInRoom = computed(() => websocketStore.usersInRoom);
const roomId = computed(() => websocketStore.roomId);
const userRole = computed(() => websocketStore.userRole);
const roundState = computed(() => websocketStore.roundState);
const ticketToGuess = computed(() => websocketStore.ticketToGuess);
const guess = computed(() => websocketStore.guess);
const didSkip = computed(() => websocketStore.didSkip);
const showAllGuesses = computed(() => websocketStore.showAllGuesses);
const possibleGuesses = computed(() => websocketStore.possibleGuesses);
const permissions = computed(() => websocketStore.permissions);
const roomIsLocked = computed(() => websocketStore.roomIsLocked);
const developerDone = computed(() => websocketStore.developerDone);

const roundIsWaiting = computed(() => roundState.value === RoundState.Waiting);

function sendMessage(
  type: SendableWebsocketMessageType,
  data: string | number | null | { password?: string; key: string },
) {
  websocketStore.send({ type, data });
}

function lockRoom() {
  showSetRoomPasswordDialog.value = false;
  sendMessage("lock-room", {
    password: roomPassword.value,
    key: permissions.value.room.key || "",
  });
}

function openRoom() {
  sendMessage("open-room", {
    key: permissions.value.room.key || "",
  });
}

function leaveRoom() {
  websocketStore.disconnect();
  router.push("/");
}

async function writeToClipboard(text: string) {
  // @ts-ignore
  const clipboardPermission = await navigator.permissions.query({ name: "clipboard-write" });
  if (clipboardPermission.state === "granted") {
    await navigator.clipboard.writeText(text);
    snackbarText.value = "Kopiert!";
  } else {
    snackbarText.value = "Konnte nicht kopiert werden";
  }
  showSnackbar.value = true;
}

async function copyPassword() {
  await writeToClipboard(roomPassword.value);
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
  <div>
    <v-dialog
      v-model="showSetRoomPasswordDialog"
      max-width="500"
    >
      <v-card>
        <v-card-title>Passwort setzen</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="roomPassword"
            placeholder="Passwort"
            :type="showPassword ? 'text' : 'password'"
            :append-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
            @click:append="showPassword = !showPassword"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            color="red"
            @click="showSetRoomPasswordDialog = false"
          >
            Abbrechen
          </v-btn>
          <v-btn
            :disabled="roomPassword.length === 0"
            color="green"
            @click="lockRoom"
          >
            Abschließen
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-toolbar rounded>
      <v-toolbar-title>
        {{ roomIsLocked ? "privater" : "öffentlicher" }} Raum: {{ roomId }}
      </v-toolbar-title>
      <div v-if="roundIsWaiting">
        <v-btn
          v-if="permissions.room.canLock && roomIsLocked"
          @click="copyPassword"
        >
          <v-tooltip
            activator="parent"
            location="bottom"
          >
            Passwort kopieren
          </v-tooltip>
          <v-icon>mdi-content-copy</v-icon>
        </v-btn>

        <v-btn
          v-if="permissions.room.canLock && roomIsLocked"
          @click="openRoom"
        >
          <v-tooltip
            activator="parent"
            location="bottom"
          >
            Raum öffnen
          </v-tooltip>
          <v-icon>mdi-key</v-icon>
        </v-btn>

        <v-btn
          v-if="permissions.room.canLock && !roomIsLocked"
          @click="showSetRoomPasswordDialog = true"
        >
          <v-tooltip
            activator="parent"
            location="bottom"
          >
            Raum schließen
          </v-tooltip>
          <v-icon>mdi-lock</v-icon>
        </v-btn>

        <v-btn @click="leaveRoom">
          <v-tooltip
            activator="parent"
            location="bottom"
          >
            Raum verlassen
          </v-tooltip>
          <v-icon>mdi-location-exit</v-icon>
        </v-btn>
      </div>
    </v-toolbar>

    <room-detail
      :users-in-room="usersInRoom"
      :developer-done="developerDone"
      :user-role="userRole"
      :round-state="roundState"
      :ticket-to-guess="ticketToGuess"
      :guess="guess"
      :did-skip="didSkip"
      :show-all-guesses="showAllGuesses"
      :possible-guesses="possibleGuesses"
      @estimate="sendMessage('estimate', $event)"
      @guess="sendMessage('guess', $event)"
      @reveal="sendMessage('reveal', null)"
      @new-round="sendMessage('new-round', null)"
      @skip="sendMessage('skip', null)"
    />

    <v-snackbar
      v-model="showSnackbar"
      :timeout="3000"
    >
      {{ snackbarText }}
    </v-snackbar>
  </div>
</template>

<style scoped></style>
