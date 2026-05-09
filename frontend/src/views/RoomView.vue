<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import RoomDetail from "@/components/RoomDetail.vue";
import { computed, onMounted, ref } from "vue";
import RoomForm from "@/components/RoomForm.vue";
import { isJust } from "@kaumlaut/pure/maybe";
import { isSuccess } from "@kaumlaut/pure/fetch-state";
import { useEstimationStore } from "@/stores/estimation.ts";
import { Role, RoundState, type SendableWebsocketMessageType } from "@/types/room.ts";

const estimationStore = useEstimationStore();
const router = useRouter();
const route = useRoute();
const showSetRoomPasswordDialog = ref(false);
const showPassword = ref(false);
const showSnackbar = ref(false);
const roomPassword = ref("");
const name = ref("");
const role = ref(Role.Empty);
const passwordForRoom = ref("");
const errorMessage = ref("");
const issueToAdd = ref("");
const showIssuesDrawer = ref(false);
const roomIsLocked = ref(false);
const queryRoomId = computed((): string => {
  if (Array.isArray(route.params.id)) {
    return route.params.id[0];
  }
  return route.params.id;
});
const possibleGuesses = computed(() => estimationStore.roomState.possibleGuesses);

const permissions = computed(() => estimationStore.roomState.permissions);
const isConnected = computed(() => estimationStore.roomState.isConnected);
const roundState = computed(() => estimationStore.roomState.roundState);
const roundIsWaiting = computed(() => estimationStore.roomState.roundState === RoundState.Waiting);

const roundStateAsReadableString = computed(() => {
  if (roundState.value === RoundState.Waiting) {
    return "Waiting for people to join…";
  } else if (
    roundState.value === RoundState.InProgress &&
    isJust(estimationStore.roomState.issueToGuess)
  ) {
    return `Currently guessing ${estimationStore.roomState.issueToGuess.value}`;
  } else {
    return "Everyone guessed!";
  }
});

function sendMessage(
  type: SendableWebsocketMessageType,
  data: string | number | null | { password?: string; key: string },
) {
  estimationStore.send({ type, data });
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
  estimationStore.leaveRoom();
  router.push("/");
}

function addIssue() {
  if (issueToAdd.value === "") {
    return;
  }

  sendMessage("add-issue", issueToAdd.value);
  issueToAdd.value = "";
}

async function writeToClipboard(text: string) {
  // @ts-ignore
  const clipboardPermission = await navigator.permissions.query({ name: "clipboard-write" });
  if (clipboardPermission.state === "granted") {
    await navigator.clipboard.writeText(text);
    estimationStore.roomNotifications.push("Copied!");
  } else {
    estimationStore.roomNotifications.push("Could not copy");
  }
  showSnackbar.value = true;
}

async function copyPassword() {
  await writeToClipboard(roomPassword.value);
}

async function tryJoin() {
  errorMessage.value = "";

  const actualRoomId = queryRoomId.value;
  const roomState = await estimationStore.fetchRoomState(actualRoomId);

  const passwordMatches = roomState.isLocked
    ? await estimationStore.authenticate(actualRoomId, passwordForRoom.value)
    : true;
  if (roomState.isLocked && !passwordMatches) {
    errorMessage.value = "The provided password is wrong";
    return;
  }

  if (roomState.inProgress) {
    errorMessage.value = "The round has already started";
    return;
  }

  const userAlreadyExistsInRoom = await estimationStore.userExists(actualRoomId, name.value);
  if (userAlreadyExistsInRoom) {
    errorMessage.value = "A user with this name already exists in the room";
    return;
  }

  await estimationStore.joinRoom(name.value, role.value, actualRoomId);
  await Promise.all([estimationStore.fetchPossibleGuesses(), estimationStore.fetchPermissions()]);
}

onMounted(async () => {
  await estimationStore
    .fetchRoomState(queryRoomId.value)
    .then((response) => {
      roomIsLocked.value = response.isLocked;
    })
    .catch(async () => {
      await router.push("/");
    });

  if (isConnected.value) {
    await Promise.all([estimationStore.fetchPossibleGuesses(), estimationStore.fetchPermissions()]);
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
      subtitle="You are currently not connected to this room"
      @submit="tryJoin"
    />
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
        {{ roundStateAsReadableString }}
      </v-toolbar-title>
      <div v-if="roundIsWaiting">
        <v-btn
          v-if="
            isJust(estimationStore.roomState.role) &&
              estimationStore.roomState.role.value == Role.ProductOwner
          "
          @click="showIssuesDrawer = !showIssuesDrawer"
        >
          <v-tooltip
            activator="parent"
            location="bottom"
          >
            Issues
          </v-tooltip>
          <v-icon>mdi-text-box-outline</v-icon>
        </v-btn>
        <v-btn
          v-if="permissions.room.canLock && estimationStore.roomState.roomIsLocked"
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
          v-if="permissions.room.canLock && estimationStore.roomState.roomIsLocked"
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
          v-if="permissions.room.canLock && !estimationStore.roomState.roomIsLocked"
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
      :room-state="estimationStore.roomState"
      @estimate="sendMessage('estimate', $event)"
      @guess="sendMessage('guess', $event)"
      @reveal="sendMessage('reveal', null)"
      @new-round="sendMessage('new-round', null)"
      @skip="sendMessage('skip', null)"
    />

    <v-navigation-drawer
      v-model="showIssuesDrawer"
      width="400"
      :location="$vuetify.display.mobile ? 'bottom' : 'right'"
    >
      <template #prepend>
        <v-container>
          <v-col>
            <v-row>
              <v-list-item title="Issues" />
              <v-spacer />
              <v-btn
                icon="mdi-close"
                variant="flat"
                size="small"
                @click="showIssuesDrawer = false"
              />
            </v-row>
          </v-col>
        </v-container>

        <v-divider />
      </template>
      <v-list>
        <v-list-item
          v-for="issue in estimationStore.roomState.issues"
          :key="issue"
        >
          <v-card
            variant="tonal"
            class="pa-2"
          >
            <v-card-title>{{ issue }}</v-card-title>

            <v-card-actions>
              <v-btn variant="tonal">
                Vote this issue
              </v-btn>
              <v-spacer />
              <span>-</span>
            </v-card-actions>
          </v-card>
        </v-list-item>
      </v-list>

      <template #append>
        <v-container>
          <v-card>
            <v-card-text>
              <v-text-field
                v-model.trim="issueToAdd"
                variant="outlined"
                placeholder="New issue"
              />

              <v-card-actions>
                <v-spacer />
                <v-btn
                  variant="outlined"
                  color="green-darken-2"
                  @click="addIssue"
                >
                  Add Issue
                </v-btn>
              </v-card-actions>
            </v-card-text>
          </v-card>
        </v-container>
      </template>
    </v-navigation-drawer>

    <v-snackbar-queue
      v-model="estimationStore.roomNotifications"
      :timeout="1500"
      color="gray"
    />
  </div>
</template>

<style scoped></style>
