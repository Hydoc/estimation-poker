<script setup lang="ts">
import ResultTable from "@/components/ResultTable.vue";
import type { Developer } from "@/components/types";

type Props = {
  ticketToGuess: string;
  showAllGuesses: boolean;
  developerList: Developer[];
  roundIsFinished: boolean;
  userIsProductOwner: boolean;
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "reveal"): void;
  (e: "new-round"): void;
}>();
</script>

<template>
  <v-card :title="`Aktuelles Ticket zum schätzen: ${props.ticketToGuess}`">
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
