<script setup lang="ts">
import { onBeforeMount, type Ref, ref } from "vue";
import RoomDialog from "@/components/RoomDialog.vue";
import { useRouter } from "vue-router";
import { isSuccess } from "@kaumlaut/pure/fetch-state";
import { useEstimationStore } from "@/stores/estimation.ts";
import { Role } from "@/types/room.ts";

const estimationStore = useEstimationStore();
// leave room when component renders to avoid weird behavior
estimationStore.leaveRoom();

const router = useRouter();
const errorMessage: Ref<string | undefined> = ref();
const role: Ref<Role> = ref(Role.Empty);
const name: Ref<string> = ref("");
const passwordForRoom: Ref<string> = ref("");
const showPasswordInput: Ref<boolean> = ref(false);

async function connect(chosenRoomId: string | undefined) {
  errorMessage.value = "";

  const actualRoomId = chosenRoomId ? chosenRoomId : await estimationStore.createRoom(name.value);
  const roomState = await estimationStore.fetchRoomState(actualRoomId);

  if (roomState.isLocked && passwordForRoom.value === "") {
    showPasswordInput.value = true;
    return;
  }

  const passwordMatches = estimationStore.roomState.roomIsLocked
    ? await estimationStore.authenticate(actualRoomId, passwordForRoom.value)
    : true;
  if (estimationStore.roomState.roomIsLocked && !passwordMatches) {
    showPasswordInput.value = true;
    errorMessage.value = "The provided password is wrong";
    return;
  }

  showPasswordInput.value = false;

  if (estimationStore.roomState.roundInProgress) {
    errorMessage.value = "The round has already started";
    return;
  }

  const userAlreadyExistsInRoom = await estimationStore.userExists(actualRoomId, name.value);
  if (userAlreadyExistsInRoom) {
    errorMessage.value = "A user with this name already exists in the room";
    return;
  }

  await estimationStore.joinRoom(name.value, role.value, actualRoomId);
  await router.push(`/room/${actualRoomId}`);
}

function playerCountAsStringForRoom(playerCount: number): string {
  return `${playerCount} player${playerCount > 1 ? "s" : ""}`;
}

onBeforeMount(async () => {
  await estimationStore.fetchActiveRooms();
});
</script>

<template>
  <main>
    <v-container>
      <div
        v-if="
          isSuccess(estimationStore.roomsState.availableActiveRooms) &&
            estimationStore.roomsState.availableActiveRooms.data.rooms.length > 0
        "
        class="d-flex flex-column"
      >
        <div class="align-self-end">
          <room-dialog
            v-if="estimationStore.roomsState.availableActiveRooms.data.rooms.length > 0"
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
            v-for="(room, index) in estimationStore.roomsState.availableActiveRooms.data.rooms"
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
