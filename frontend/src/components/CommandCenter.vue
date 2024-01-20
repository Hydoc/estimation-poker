<script setup lang="ts">
import { Role, RoundState } from "@/components/types";
import DeveloperCommandCenter from "@/components/DeveloperCommandCenter.vue";
import { computed } from "vue";
import ProductOwnerCommandCenter from "@/components/ProductOwnerCommandCenter.vue";

type Props = {
  userRole: Role;
  roundState: RoundState;
  guess: number;
  ticketToGuess: string;
  hasDevelopersInRoom: boolean;
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "guess", guess: number): void;
}>();

const isDeveloper = computed(() => props.userRole === Role.Developer);
const roundIsWaiting = computed(() => props.roundState === RoundState.Waiting);
const didGuess = computed(() => props.guess !== 0);
const hasTicketToGuess = computed(() => props.ticketToGuess !== "");
</script>

<template>
  <developer-command-center
    v-if="isDeveloper"
    :did-guess="didGuess"
    :has-ticket-to-guess="hasTicketToGuess"
    @guess="emit('guess', $event)"
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
