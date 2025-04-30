<script setup lang="ts">
import { computed } from "vue";
import type { UserOverview } from "./types";

type Props = {
  usersInRoom: UserOverview;
};
const props = defineProps<Props>();
const radius = 200;
const cy = 300;
const cx = 300;
const users = computed(() => {
  return [...props.usersInRoom.productOwnerList, ...props.usersInRoom.developerList];
});

function topForElement(index: number): string {
  const theta = 2 * Math.PI * (index / users.value.length);
  const top = cy - radius * Math.cos(theta);
  return `${top}px`;
}

function leftForElement(index: number): string {
  const theta = 2 * Math.PI * (index / users.value.length);
  const left = cx + radius * Math.sin(theta);
  return `${left}px`;
}
</script>

<template>
  <ul class="virtual-table">
    <li
      v-for="(user, index) in users"
      :key="user.name"
      :style="`left:${leftForElement(index)};top:${topForElement(index)}`"
    >
      {{ user.name }}
    </li>
  </ul>
</template>

<style scoped>
.virtual-table {
  position: relative;
  width: 600px;
  height: 600px;
}

.virtual-table > li {
  position: absolute;
  list-style: none;
}
</style>
