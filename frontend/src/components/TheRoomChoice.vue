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

function connect() {
  websocketStore.connect(name.value, role.value, roomId.value);
  router.push("/room");
}

const isButtonEnabled = computed(() => roomId.value !== "" && name.value !== "" && role.value !== "");

</script>

<template>
  <v-container>
    <v-row align="center" justify="center">
      <v-col>
        <v-card prepend-icon="mdi-poker-chip">
          <template #title>
            Ich brauche noch ein paar Informationen bevor es los geht
          </template>
          <v-card-text>

            <v-container>

              <v-form ref="form" fast-fail @submit.prevent="connect" validate-on="input">
                <v-text-field label="Raum ID" v-model="roomId" required :rules="[v => !!v || 'Fehler: Hier müsste eigentlich was stehen']" />
                <v-text-field label="Name" v-model="name" required :rules="[v => !!v || 'Fehler: Hier müsste eigentlich was stehen']" />

                <v-radio-group label="Deine Rolle" v-model="role">
                  <v-radio label="Product Owner" :value="Role.ProductOwner"></v-radio>
                  <v-radio label="Entwickler" :value="Role.Developer"></v-radio>
                </v-radio-group>

                <v-col class="text-right">
                  <v-btn type="submit" color="primary" prepend-icon="mdi-connection" class="mx-auto" :disabled="!isButtonEnabled">Verbinden</v-btn>
                </v-col>
              </v-form>
            </v-container>

          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>

</style>