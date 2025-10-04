<script setup lang="ts">
import { useWebsocketStore } from "@/stores/websocket";
import { onBeforeMount, type Ref, ref } from "vue";
import RoomDialog from "@/components/RoomDialog.vue";
import { type ActiveRoom, Role } from "@/components/types.ts";
import { useRouter } from "vue-router";
const websocketStore = useWebsocketStore();
websocketStore.disconnect();
websocketStore.resetRound();

const router = useRouter();
const activeRooms: Ref<ActiveRoom[]> = ref([]);
const errorMessage: Ref<string | undefined> = ref();
const role: Ref<Role> = ref(Role.Empty);
const name: Ref<string> = ref("");
const passwordForRoom: Ref<string> = ref("");
const showPasswordInput: Ref<boolean> = ref(false);

async function connect(chosenRoomId: string | undefined) {
  errorMessage.value = "";

  const actualRoomId = chosenRoomId ? chosenRoomId : await websocketStore.createRoom(name.value);

  const isLocked = await websocketStore.isRoomLocked(actualRoomId);
  if (isLocked && passwordForRoom.value === "") {
    showPasswordInput.value = true;
    return;
  }

  const passwordMatches = isLocked
    ? await websocketStore.passwordMatchesRoom(actualRoomId, passwordForRoom.value)
    : true;
  if (isLocked && !passwordMatches) {
    showPasswordInput.value = true;
    errorMessage.value = "The provided password is wrong";
    return;
  }

  showPasswordInput.value = false;

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
  await router.push(`/room/${actualRoomId}`);
}

function playerCountAsStringForRoom(playerCount: number): string {
  return `${playerCount} player${playerCount > 1 ? "s" : ""}`;
}

onBeforeMount(async () => {
  activeRooms.value = await websocketStore.fetchActiveRooms();
});
</script>

<template>
  <main>
    <v-container>
      <div
        v-if="activeRooms.length > 0"
        class="d-flex flex-column"
      >
        <div class="align-self-end">
          <room-dialog
            v-if="activeRooms.length > 0"
            v-model:role="role"
            v-model:name="name"
            activator-text="Create a new room"
            card-title="Create room"
            :error-message="errorMessage"
            @submit="connect(undefined)"
          />
        </div>
        <div class="d-flex ga-5 flex-wrap">
          <v-card
            v-for="(room, index) in activeRooms"
            :key="room.id"
            variant="outlined"
            prepend-icon="mdi-poker-chip"
            max-width="450"
            :title="`Room #${index + 1}`"
            :subtitle="room.id"
          >
            <v-card-text>
              <v-icon icon="mdi-account" />
              {{ playerCountAsStringForRoom(room.playerCount) }}
            </v-card-text>
            <v-card-actions>
              <v-spacer />

              <room-dialog
                v-model:role="role"
                v-model:name="name"
                v-model:password="passwordForRoom"
                :show-password-input="showPasswordInput"
                activator-text="Join"
                card-title="Join"
                :error-message="errorMessage"
                @submit="connect(room.id)"
              />
            </v-card-actions>
          </v-card>
        </div>
      </div>

      <div
        v-else
        class="d-flex align-center flex-column ga-7"
      >
        <v-icon
          icon="mdi-magnify"
          class="opacity-50"
          size="80"
        />

        <span class="text-h4 opacity-90">There are currently no rooms</span>

        <room-dialog
          v-model:role="role"
          v-model:name="name"
          activator-text="Create a new one"
          card-title="Create room"
          :error-message="errorMessage"
          @submit="connect(undefined)"
        />
      </div>
    </v-container>
  </main>
</template>
