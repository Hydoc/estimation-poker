<script setup lang="ts">
import type { Developer } from "@/components/types";
import { computed } from "vue";

type Props = {
  developerList: Developer[];
  showAllGuesses: boolean;
  roundIsFinished: boolean;
};

const props = defineProps<Props>();

const averageGuess = computed(() => {
  return Math.round(
    props.developerList.reduce((sum, dev) => sum + dev.guess, 0) / props.developerList.length,
  );
});
</script>

<template>
  <v-table>
    <thead>
      <tr>
        <th>Name</th>
        <th>Schätzung</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="developer in props.developerList" :key="developer.name" :class="{'bg-blue-grey-lighten-5': developer.guess !== averageGuess && props.showAllGuesses}">
        <td>{{ developer.name }}</td>
        <td>
          <v-icon color="green" v-if="developer.guess !== 0 && !props.showAllGuesses"
            >mdi-check-circle</v-icon
          >
          <v-icon v-else-if="developer.guess === 0">mdi-help-circle</v-icon>
          <span v-if="props.showAllGuesses">{{ developer.guess }}</span>
        </td>
      </tr>
      <tr v-if="props.showAllGuesses">
        <td class="font-weight-500">Durchschnitt</td>
        <td class="font-weight-500">{{ averageGuess }}</td>
      </tr>
    </tbody>
  </v-table>
</template>

<style scoped>
.font-weight-500 {
  font-weight: 500;
}
</style>
