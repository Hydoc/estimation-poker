<script setup lang="ts">
import type { Developer, DeveloperDone } from "@/components/types";
import { computed } from "vue";

type Props = {
  developerList: Developer[];
  developerDone: DeveloperDone[];
  showAllGuesses: boolean;
  roundIsFinished: boolean;
};

const props = defineProps<Props>();

const averageGuess = computed(() => {
  const developerThatDoNotSkip = props.developerDone.filter((dev) => !dev.doSkip);
  const average = Math.round(
    developerThatDoNotSkip.reduce((sum, dev) => sum + dev.guess, 0) / developerThatDoNotSkip.length,
  );
  return Number.isNaN(average) ? 0 : average;
});

function developerDidNotGuessAverage(dev: Developer): boolean {
  const foundDeveloperDone = foundDeveloper(dev);
  return (
    foundDeveloperDone?.guess !== averageGuess.value &&
    props.showAllGuesses &&
    !foundDeveloperDone?.doSkip
  );
}

function foundDeveloper(dev: Developer): DeveloperDone | null {
  return props.developerDone.find((it) => it.name === dev.name) || null;
}
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
      <tr
        v-for="developer in props.developerList"
        :key="developer.name"
        :class="{
          'bg-blue-grey-lighten-5': developerDidNotGuessAverage(developer),
        }"
      >
        <td>{{ developer.name }}</td>
        <td>
          <span v-if="!props.showAllGuesses">
            <v-icon
              v-if="developer.isDone"
              color="green"
            >mdi-check-circle</v-icon>
            <v-icon v-else>mdi-help-circle</v-icon>
          </span>

          <span v-else>
            <span v-if="!foundDeveloper(developer)?.doSkip">{{
              foundDeveloper(developer)?.guess
            }}</span>
            <span v-else><v-icon>mdi-coffee</v-icon></span>
          </span>
        </td>
      </tr>
      <tr v-if="props.showAllGuesses">
        <td class="font-weight-500">
          Durchschnitt
        </td>
        <td class="font-weight-500">
          {{ averageGuess }}
        </td>
      </tr>
    </tbody>
  </v-table>
</template>

<style scoped>
.font-weight-500 {
  font-weight: 500;
}
</style>
