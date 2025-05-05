<script setup lang="ts">
import {
  type DeveloperDone,
  type PossibleGuess,
  Role,
  RoundState,
  type UserOverview,
} from "@/components/types";
import { computed, ref, watch } from "vue";
import TableOverview from "@/components/TableOverview.vue";
import DeveloperCommandCenter from "@/components/DeveloperCommandCenter.vue";
import RoundSummary from "@/components/RoundSummary.vue";

type Props = {
  usersInRoom: UserOverview;
  developerDone: DeveloperDone[];
  userRole: Role;
  roundState: RoundState;
  ticketToGuess: string;
  guess: number;
  didSkip: boolean;
  showAllGuesses: boolean;
  possibleGuesses: PossibleGuess[];
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "guess", guess: number): void;
  (e: "reveal"): void;
  (e: "new-round"): void;
  (e: "skip"): void;
}>();
const showRoundSummary = ref(false);
const delay = 500;

const userIsDeveloper = computed(() => props.userRole === Role.Developer);
const hasTicketToGuess = computed(() => props.ticketToGuess !== "");

watch(
  () => props.showAllGuesses,
  (doShowAllGuesses: boolean) => {
    if (doShowAllGuesses) {
      setTimeout(() => {
        showRoundSummary.value = true;
      }, delay);
    } else {
      showRoundSummary.value = false;
    }
  },
);
</script>

<template>
  <round-summary
    v-if="showRoundSummary"
    :developer-done="props.developerDone"
  />

  <v-container fluid>
    <v-col cols="12">
      <v-row>
        <table-overview
          :show-all-guesses="props.showAllGuesses"
          :users-in-room="props.usersInRoom"
          :round-state="props.roundState"
          :user-role="props.userRole"
          :developer-done="developerDone"
          :ticket-to-guess="props.ticketToGuess"
          @estimate="emit('estimate', $event)"
          @reveal="emit('reveal')"
          @new-round="emit('new-round')"
        />
      </v-row>

      <v-row
        align="center"
        justify="center"
      >
        <developer-command-center
          v-if="userIsDeveloper"
          class="developer-command-center"
          :show-all-guesses="props.showAllGuesses"
          :guess="props.guess"
          :did-skip="props.didSkip"
          :has-ticket-to-guess="hasTicketToGuess"
          :possible-guesses="props.possibleGuesses"
          @guess="emit('guess', $event)"
          @skip="emit('skip')"
        />
      </v-row>
    </v-col>
  </v-container>
</template>

<style scoped>
.developer-command-center {
  margin-left: 2rem;
  margin-top: 5rem;
}
</style>
