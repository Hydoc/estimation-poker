<script setup lang="ts">
import ResultTable from "@/components/ResultTable.vue";
import type { Developer } from "@/components/types";
import { computed } from "vue";

type Props = {
  ticketToGuess: string;
  showAllGuesses: boolean;
  developerList: Developer[];
  roundIsFinished: boolean;
  userIsProductOwner: boolean;
};
const percentageDone = computed(() => {
  const devsThatAreDone = props.developerList.filter((dev) => dev.doSkip || dev.guess > 0).length;
  const totalDevs = props.developerList.length;
  return Math.round((devsThatAreDone / totalDevs) * 100);
});

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "reveal"): void;
  (e: "new-round"): void;
}>();
</script>

<template>
  <v-card>
    <v-card-title class="d-flex justify-space-between">
      <span>Aktuelles Ticket zum schätzen: {{ props.ticketToGuess }}</span>
      <v-progress-circular :model-value="percentageDone" size="50" width="5" color="teal-darken-1">
        <template #default>
          <span class="text-body-2"> {{ percentageDone }}% </span>
        </template>
      </v-progress-circular>
    </v-card-title>
    <v-container>
      <result-table
        :developer-list="props.developerList"
        :show-all-guesses="props.showAllGuesses"
        :round-is-finished="props.roundIsFinished"
      />

      <v-card-actions v-if="props.roundIsFinished && props.userIsProductOwner">
        <v-spacer />
        <v-btn v-if="!props.showAllGuesses" color="primary" @click="emit('reveal')">Auflösen</v-btn>
        <v-btn v-if="props.showAllGuesses" color="blue-darken-4" @click="emit('new-round')"
          >Neue Runde</v-btn
        >
      </v-card-actions>
    </v-container>
  </v-card>
</template>

<style scoped></style>
