<script setup lang="ts">
import type { UserOverview } from "@/components/types";
import UserBox from "@/components/UserBox.vue";

type Props = {
  roomId: string;
  usersInRoom: UserOverview;
  currentUsername: string;
};

const props = defineProps<Props>();
</script>

<template>
  <h1>Raum: {{ props.roomId }}</h1>

  <v-container width="200">
    <v-row>
      <v-col>
        <UserBox
          title="Product Owner"
          :user-list="usersInRoom.productOwnerList"
          :current-username="currentUsername"
        />
      </v-col>
      <v-col>
        <UserBox
          title="Entwickler"
          :user-list="usersInRoom.developerList"
          :current-username="currentUsername"
        />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
