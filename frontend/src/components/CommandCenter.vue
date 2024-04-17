<script setup lang="ts">
import { type PossibleGuess, Role, RoundState } from "@/components/types";
import DeveloperCommandCenter from "@/components/DeveloperCommandCenter.vue";
import { computed } from "vue";
import ProductOwnerCommandCenter from "@/components/ProductOwnerCommandCenter.vue";

type Props = {
  userRole: Role;
  roundState: RoundState;
  guess: number;
  showAllGuesses: boolean;
  didSkip: boolean;
  ticketToGuess: string;
  hasDevelopersInRoom: boolean;
  possibleGuesses: PossibleGuess[];
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "guess", guess: number): void;
  (e: "skip"): void;
}>();

const isDeveloper = computed(() => props.userRole === Role.Developer);
const roundIsWaiting = computed(() => props.roundState === RoundState.Waiting);
const hasTicketToGuess = computed(() => props.ticketToGuess !== "");
</script>

<template>
  <developer-command-center
    v-if="isDeveloper"
    :show-all-guesses="props.showAllGuesses"
    :guess="props.guess"
    :did-skip="props.didSkip"
    :has-ticket-to-guess="hasTicketToGuess"
    :possible-guesses="props.possibleGuesses"
    @guess="emit('guess', $event)"
    @skip="emit('skip')"
  />
  <product-owner-command-center
    v-else
    :round-is-waiting="roundIsWaiting"
    :has-ticket-to-guess="hasTicketToGuess"
    :has-developers-in-room="props.hasDevelopersInRoom"
    @estimate="emit('estimate', $event)"
  />
</template>

<style scoped></style>
