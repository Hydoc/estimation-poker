<script setup lang="ts">
import type { User } from "@/components/types";

type Props = {
  userList: User[];
  currentUsername: string;
  title: string;
};

const props = defineProps<Props>();

function isThisUser(user: User) {
  return user.name === props.currentUsername;
}

function formatUser(user: User): string {
  return `${user.name}${isThisUser(user) ? ` (Du)` : ""}`;
}
</script>

<template>
  <v-card>
    <v-card-title>{{ props.title }}</v-card-title>
    <v-card-text>
      <v-list v-if="props.userList.length > 0">
        <v-list-item v-for="user in props.userList" :key="user.name">
          <v-list-item-title>{{ formatUser(user) }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<style scoped></style>
