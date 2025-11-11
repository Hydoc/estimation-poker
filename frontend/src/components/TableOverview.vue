<script setup lang="ts">
import { computed } from "vue";
import {
  type Developer,
  type DeveloperDone,
  type ProductOwner,
  Role,
  RoundState,
  type UserOverview,
} from "./types";
import DeveloperCard from "@/components/DeveloperCard.vue";
import ProductOwnerCommandCenter from "@/components/ProductOwnerCommandCenter.vue";

type Props = {
  usersInRoom: UserOverview;
  roundState: RoundState;
  developerDone: DeveloperDone[];
  showAllGuesses: boolean;
  userRole: Role;
  ticketToGuess: string;
};
const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "reveal"): void;
  (e: "new-round"): void;
  (e: "estimate", ticket: string): void;
}>();
const radius = 250;
const cy = 300;
const cx = 300;
const developerList = computed(() => props.usersInRoom.filter((it) => it.role === "developer"));

const userIsProductOwner = computed(() => props.userRole === Role.ProductOwner);
const userIsDeveloper = computed(() => props.userRole === Role.Developer);
const hasTicketToGuess = computed(() => props.ticketToGuess !== "");

function isDeveloper(user: ProductOwner | Developer): user is Developer {
  return user.role === "developer";
}

function findDeveloperDone(developer: Developer): DeveloperDone | undefined {
  return props.developerDone.find((it) => it.name === developer.name);
}

function topForElement(index: number): string {
  const theta = 2 * Math.PI * (index / props.usersInRoom.length);
  const top = cy - radius * Math.cos(theta);
  return `${top}px`;
}

function leftForElement(index: number, username: string): string {
  const theta = 2 * Math.PI * (index / props.usersInRoom.length);
  const left = cx + radius * Math.sin(theta);
  return `${left - (username.length * 2 + 1)}px`;
}
</script>

<template>
  <div class="virtual-table">
    <div class="table">
      <product-owner-command-center
        v-if="userIsProductOwner"
        :round-state="props.roundState"
        :developer-list="developerList"
        :actual-ticket-to-guess="props.ticketToGuess"
        :has-ticket-to-guess="hasTicketToGuess"
        :show-all-guesses="props.showAllGuesses"
        @estimate="emit('estimate', $event)"
        @reveal="emit('reveal')"
        @new-round="emit('new-round')"
      />
      <span v-if="!hasTicketToGuess && !userIsProductOwner">Waiting for ticket…</span>
      <span
        v-if="hasTicketToGuess && userIsDeveloper"
        class="text-h5"
      >{{
        props.ticketToGuess
      }}</span>
    </div>
    <div
      v-for="(user, index) in props.usersInRoom"
      :key="user.name"
      class="seat"
      :style="`left:${leftForElement(index, user.name)};top:${topForElement(index)}`"
    >
      <developer-card
        v-if="isDeveloper(user)"
        :developer="user"
        :developer-done="findDeveloperDone(user)"
      />
      <span v-else>{{ user.name }}</span>
    </div>
  </div>
</template>

<style scoped>
.virtual-table {
  position: relative;
  margin: 0 auto;
  width: 37.5rem;
  height: 37.5rem;
}

.seat {
  position: absolute;
}

.table {
  box-shadow: rgba(100, 100, 111, 0.2) 0 7px 29px 0;
  position: absolute;
  top: 25%;
  left: 8.5rem;
  width: 350px;
  height: 350px;
  background-color: #d7e9ff;
  border-radius: 50%;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  justify-content: space-evenly;
  align-items: center;
}
</style>
