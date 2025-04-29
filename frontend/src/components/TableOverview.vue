<script setup lang="ts">
import { type Ref, onBeforeMount, ref } from 'vue';
import type { Developer } from './types';

type Props = {
  developerList: Developer[];
}
const props = defineProps<Props>();
const seats: Ref<HTMLLIElement[] | null> = ref(null);

onBeforeMount(() => {
  const cx = 300;
  const cy = 300;
  const radius = 200;

  seats.value?.forEach((seat: HTMLLIElement, index: number, ) => {
    const theta = 2 * Math.PI * (index / seats.value!.length);
    const left = cx + radius * Math.sin(theta);
    const top = cy - radius * Math.cos(theta);
    seat.style.left = left.toString() + "px";
    seat.style.top = top.toString() + "px";
  });
});
</script>

<template>
  <ul class="virtual-table">
    <li ref="seats" v-for="developer in props.developerList" :key="developer.name">{{ developer.name }}</li>
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
