<script setup lang="ts">
import { useWebsocketStore } from "@/stores/websocket";
import { useRouter } from "vue-router";
import { computed } from "vue";
import type { Developer, ProductOwner } from "@/components/types";

const websocketStore = useWebsocketStore();
const router = useRouter();
if (!websocketStore.isConnected) {
  router.push("/");
}
const usersInRoom = computed(() => websocketStore.usersInRoom);
const roomId = computed(() => websocketStore.roomId);

function isThisUser(user: ProductOwner | Developer) {
  return user.name === websocketStore.username;
}

function formatUser(user: ProductOwner | Developer): string {
  return `${user.name}${isThisUser(user) ? ` (Du)` : ""}`;
}
</script>

<template>
  <h1>Raum: {{ roomId }}</h1>

  <v-container width="200">
    <v-row>
      <v-col>
        <v-card>
          <v-card-title>Product Owner</v-card-title>
          <v-card-text>
            <ul v-if="usersInRoom.productOwnerList.length > 0">
              <li v-for="user in usersInRoom.productOwnerList" :key="user.name">
                {{ formatUser(user) }}
              </li>
            </ul>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col>
        <v-card>
          <v-card-title>Entwickler</v-card-title>
          <v-card-text>
            <ul v-if="usersInRoom.developerList.length > 0">
              <li v-for="user in usersInRoom.developerList" :key="user.name">
                {{ formatUser(user) }}
              </li>
            </ul>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
