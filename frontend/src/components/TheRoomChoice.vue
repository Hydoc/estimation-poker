<script setup lang="ts">
import { computed, ref } from "vue";
import type { Ref } from "vue";
import { useWebsocketStore } from "@/stores/websocket";
import { Role } from "@/components/types";
import { useRouter } from "vue-router";

const router = useRouter();
const roomId = ref("");
const name = ref("");
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

const isButtonEnabled = computed(
  () => roomId.value !== "" && name.value !== "" && role.value !== "",
);

const textFieldRules = computed(() => [
  (value: string) => !!value || "Fehler: Hier müsste eigentlich was stehen",
]);
</script>

<template>
  <v-container>
    <v-row align="center" justify="center">
      <v-col>
        <v-card prepend-icon="mdi-poker-chip">
          <template #title> Ich brauche noch ein paar Informationen bevor es los geht </template>
          <v-card-text>
            <v-container>
              <v-form :fast-fail="true" @submit.prevent="connect" validate-on="input">
                <v-col>
                  <v-text-field label="Raum ID" v-model="roomId" required :rules="textFieldRules" />
                  <v-text-field label="Name" v-model="name" required :rules="textFieldRules" />
                </v-col>

                <v-radio-group label="Deine Rolle" v-model="role">
                  <v-radio label="Product Owner" :value="Role.ProductOwner"></v-radio>
                  <v-radio label="Entwickler" :value="Role.Developer"></v-radio>
                </v-radio-group>

                <v-col v-if="errorMessage !== ''">
                  <v-alert color="error" :text="errorMessage" />
                </v-col>

                <v-col class="text-right">
                  <v-btn
                    type="submit"
                    color="primary"
                    prepend-icon="mdi-connection"
                    class="mx-auto"
                    :disabled="!isButtonEnabled"
                    >Verbinden</v-btn
                  >
                </v-col>
              </v-form>
            </v-container>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
