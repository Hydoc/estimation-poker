<script setup lang="ts">
import {
  type DeveloperDone,
  type Permissions,
  type PossibleGuess,
  Role,
  RoundState,
  type UserOverview,
} from "@/components/types";
import UserBox from "@/components/UserBox.vue";
import CommandCenter from "@/components/CommandCenter.vue";
import { computed, ref } from "vue";
import TableOverview from "@/components/TableOverview.vue";

type Props = {
  roomId: string;
  usersInRoom: UserOverview;
  developerDone: DeveloperDone[];
  currentUsername: string;
  userRole: Role;
  roundState: RoundState;
  ticketToGuess: string;
  guess: number;
  didSkip: boolean;
  showAllGuesses: boolean;
  possibleGuesses: PossibleGuess[];
  permissions: Permissions;
  roomIsLocked: boolean;
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "guess", guess: number): void;
  (e: "reveal"): void;
  (e: "new-round"): void;
  (e: "leave"): void;
  (e: "skip"): void;
  (e: "lock-room", payload: { password: string; key: string }): void;
  (e: "open-room", payload: { key: string }): void;
}>();
const showSnackbar = ref(false);
const snackbarText = ref("");
const showSetRoomPasswordDialog = ref(false);
const roomPassword = ref("");
const showPassword = ref(false);
const roundIsFinished = computed(() => props.roundState === RoundState.End);
const userIsProductOwner = computed(() => props.userRole === Role.ProductOwner);
const roundIsWaiting = computed(() => props.roundState === RoundState.Waiting);
const roomIsLockedText = computed(() => (props.roomIsLocked ? "privater" : "öffentlicher"));

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

async function copyRoomName() {
  await writeToClipboard(props.roomId);
}

async function copyPassword() {
  await writeToClipboard(roomPassword.value);
}

function lockRoom() {
  showSetRoomPasswordDialog.value = false;
  emit("lock-room", {
    password: roomPassword.value,
    key: props.permissions.room.key || "",
  });
}

function openRoom() {
  emit("open-room", {
    key: props.permissions.room.key || "",
  });
}
</script>

<template>
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

  <v-container>
    <v-row>
      <v-col>
        <h1>
          {{ roomIsLockedText }} Raum: {{ props.roomId }}
          <v-icon
            title="Raum kopieren"
            size="x-small"
            @click="copyRoomName"
          >
            mdi-content-copy
          </v-icon>
        </h1>
      </v-col>
      <v-col
        v-if="roundIsWaiting"
        class="text-right align-self-center"
      >
        <v-btn
          class="mr-1"
          append-icon="mdi-location-exit"
          color="deep-purple-darken-1"
          @click="emit('leave')"
        >
          Raum verlassen
        </v-btn>
        <v-btn
          v-if="permissions.room.canLock && !roomIsLocked"
          append-icon="mdi-lock"
          color="grey-darken-2"
          @click="showSetRoomPasswordDialog = true"
        >
          Raum schließen
        </v-btn>
        <v-btn
          v-if="permissions.room.canLock && roomIsLocked"
          class="mr-1"
          append-icon="mdi-key"
          color="grey-darken-2"
          @click="openRoom"
        >
          Raum öffnen
        </v-btn>
        <v-btn
          v-if="roomIsLocked && permissions.room.canLock"
          color="indigo-darken-3"
          append-icon="mdi-content-copy"
          @click="copyPassword"
        >
          Passwort kopieren
        </v-btn>
      </v-col>
    </v-row>
  </v-container>

  <v-container>
    <v-row>
      <v-col>
        <user-box
          title="Product Owner"
          :user-list="usersInRoom.productOwnerList"
          :current-username="currentUsername"
        />
      </v-col>
      <v-col>
        <user-box
          title="Entwickler"
          :user-list="usersInRoom.developerList"
          :current-username="currentUsername"
        />
      </v-col>
    </v-row>

    <table-overview :users-in-room="props.usersInRoom" />

    <v-row class="mt-15">
      <v-col cols="12">
        <command-center
          :user-role="props.userRole"
          :round-state="props.roundState"
          :guess="props.guess"
          :did-skip="props.didSkip"
          :ticket-to-guess="props.ticketToGuess"
          :has-developers-in-room="props.usersInRoom.developerList.length > 0"
          :possible-guesses="props.possibleGuesses"
          :show-all-guesses="props.showAllGuesses"
          @estimate="emit('estimate', $event)"
          @skip="emit('skip')"
          @guess="emit('guess', $event)"
        />
      </v-col>
    </v-row>
  </v-container>
  <v-snackbar
    v-model="showSnackbar"
    :timeout="3000"
  >
    {{ snackbarText }}
  </v-snackbar>
</template>

<style scoped></style>
