<script setup lang="ts">
import { Role, type UserOverview } from "@/components/types";
import UserBox from "@/components/UserBox.vue";
import CommandCenter from "@/components/CommandCenter.vue";

type Props = {
  roomId: string;
  usersInRoom: UserOverview;
  currentUsername: string;
  userRole: Role;
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

    <v-row class="mt-15">
      <v-col cols="12">
        <CommandCenter :user-role="props.userRole" />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
