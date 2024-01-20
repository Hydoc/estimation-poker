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
  developersInRoom: boolean;
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "guess", guess: number): void;
}>();

const isDeveloper = computed(() => props.userRole === Role.Developer);
const isProductOwner = computed(() => props.userRole === Role.ProductOwner);
const roundIsInProgress = computed(() => props.roundState === RoundState.InProgress);
const didGuess = computed(() => props.guess !== 0);
const hasTicketToGuess = computed(() => props.ticketToGuess !== "");
</script>

<template>
  <developer-command-center
    v-if="isDeveloper"
    :round-is-in-progress="roundIsInProgress"
    :did-guess="didGuess"
    :has-ticket-to-guess="hasTicketToGuess"
    @guess="emit('guess', $event)"
  />
  <product-owner-command-center
    :round-state="roundState"
    v-else-if="isProductOwner && !roundIsInProgress"
    :developers-in-room="props.developersInRoom"
    @estimate="emit('estimate', $event)"
  />
</template>

<style scoped></style>
