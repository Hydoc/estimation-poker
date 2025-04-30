<script setup lang="ts">
import { computed, onUpdated, type Ref, ref } from 'vue';
import type {  UserOverview } from './types';

type Props = {
  usersInRoom: UserOverview;
}
const props = defineProps<Props>();
const seats: Ref<HTMLLIElement[] | null> = ref(null);
const users = computed(() => {
  return [...props.usersInRoom.productOwnerList, ...props.usersInRoom.developerList];
});

onUpdated(() => {
  const cx = 300;
  const cy = 300;
  const radius = 200;

  seats.value!.forEach((seat: HTMLLIElement, index: number, ) => {
    const theta = 2 * Math.PI * (index / seats.value!.length);
    const left = cx + radius * Math.sin(theta);
    const top = cy - radius * Math.cos(theta);
    seat.style.left = `${left}px`;
    seat.style.top = `${top}px`;
  });
});
</script>

<template>
  <ul class="virtual-table">
    <li ref="seats" v-for="user in users" :key="user.name">{{ user.name }}</li>
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
