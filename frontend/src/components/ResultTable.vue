<script setup lang="ts">
import type { Developer } from "@/components/types";

type Props = {
  developerList: Developer[];
  showAllGuesses: boolean;
  roundIsFinished: boolean;
};

const props = defineProps<Props>();
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
      <tr v-for="developer in props.developerList" :key="developer.name">
        <td>{{ developer.name }}</td>
        <td>
          <v-icon color="green" v-if="developer.guess !== 0 && !props.showAllGuesses"
            >mdi-check-circle</v-icon
          >
          <v-icon v-else-if="developer.guess === 0">mdi-help-circle</v-icon>
          <span v-if="props.showAllGuesses">{{ developer.guess }}</span>
        </td>
      </tr>
    </tbody>
  </v-table>
</template>

<style scoped></style>
