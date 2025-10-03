<script setup lang="ts">
import { type SendableWebsocketMessageType, useWebsocketStore } from "@/stores/websocket";
import { useRoute, useRouter } from "vue-router";
import RoomDetail from "@/components/RoomDetail.vue";
import { ref, computed, onMounted } from "vue";
import { Role, RoundState } from "@/components/types.ts";
import RoomForm from "@/components/RoomForm.vue";

const websocketStore = useWebsocketStore();
const router = useRouter();
const route = useRoute();
const showSetRoomPasswordDialog = ref(false);
const showPassword = ref(false);
const showSnackbar = ref(false);
const roomPassword = ref("");
const snackbarText = ref("");
const name = ref("");
const role = ref(Role.Empty);
const passwordForRoom = ref("");
const errorMessage = ref("");
const roomIsLocked = computed(() => websocketStore.roomIsLocked);
const usersInRoom = computed(() => websocketStore.usersInRoom);
const roomId = computed(() => websocketStore.roomId);
const queryRoomId = computed((): string => {
  if (Array.isArray(route.params.id)) {
    return route.params.id[0];
  }
  return route.params.id;
});
const userRole = computed(() => websocketStore.userRole);
const roundState = computed(() => websocketStore.roundState);
const ticketToGuess = computed(() => websocketStore.ticketToGuess);
const guess = computed(() => websocketStore.guess);
const didSkip = computed(() => websocketStore.didSkip);
const showAllGuesses = computed(() => websocketStore.showAllGuesses);
const possibleGuesses = computed(() => websocketStore.possibleGuesses);
const permissions = computed(() => websocketStore.permissions);
const developerDone = computed(() => websocketStore.developerDone);
const isConnected = computed(() => websocketStore.isConnected);

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
    snackbarText.value = "Copied!";
  } else {
    snackbarText.value = "Could not copy";
  }
  showSnackbar.value = true;
}

async function copyPassword() {
  await writeToClipboard(roomPassword.value);
}

async function tryJoin() {
  errorMessage.value = "";

  const actualRoomId = queryRoomId.value;

  const passwordMatches = roomIsLocked.value
    ? await websocketStore.passwordMatchesRoom(actualRoomId, passwordForRoom.value)
    : true;
  if (roomIsLocked.value && !passwordMatches) {
    errorMessage.value = "The provided password is wrong";
    return;
  }

  const roundInRoomInProgress = await websocketStore.isRoundInRoomInProgress(actualRoomId);
  if (roundInRoomInProgress) {
    errorMessage.value = "The round has already started";
    return;
  }

  const userAlreadyExistsInRoom = await websocketStore.userExistsInRoom(name.value, actualRoomId);
  if (userAlreadyExistsInRoom) {
    errorMessage.value = "A user with this name already exists in the room";
    return;
  }

  await websocketStore.connect(name.value, role.value, actualRoomId);
  await Promise.all([websocketStore.fetchPossibleGuesses(), websocketStore.fetchPermissions()]);
}

onMounted(async () => {
  const roomExists = await websocketStore.roomExists(queryRoomId.value);
  if (!roomExists) {
    await router.push("/");
    return;
  }

  await websocketStore.fetchRoomIsLocked(queryRoomId.value);

  if (isConnected.value) {
    await Promise.all([websocketStore.fetchPossibleGuesses(), websocketStore.fetchPermissions()]);
  }
});
</script>

<template>
  <div v-if="!isConnected">
    <room-form
      v-model:name="name"
      v-model:role="role"
      v-model:password="passwordForRoom"
      :show-password-input="roomIsLocked"
      :error-message="errorMessage"
      title="Join room"
      @submit="tryJoin"
    >
      <template #teaser>
        <p>You are currently not connected to this room.</p>
      </template>
    </room-form>
  </div>

  <div v-else>
    <v-dialog
      v-model="showSetRoomPasswordDialog"
      max-width="500"
    >
      <v-card>
        <v-card-title>Set password</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="roomPassword"
            placeholder="Password"
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
            Cancel
          </v-btn>
          <v-btn
            :disabled="roomPassword.length === 0"
            color="green"
            @click="lockRoom"
          >
            Lock
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-toolbar rounded>
      <v-toolbar-title>
        {{ roomIsLocked ? "private" : "public" }} room: {{ roomId }}
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
            Copy password
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
            Unlock room
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
            Lock room
          </v-tooltip>
          <v-icon>mdi-lock</v-icon>
        </v-btn>

        <v-btn @click="leaveRoom">
          <v-tooltip
            activator="parent"
            location="bottom"
          >
            Leave room
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
