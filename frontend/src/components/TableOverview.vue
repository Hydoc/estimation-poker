<script setup lang="ts">
import { computed } from "vue";
import { type Developer, type DeveloperDone, type ProductOwner, RoundState, type UserOverview } from "./types";
import DeveloperCard from "@/components/DeveloperCard.vue";
import ProductOwnerCommandCenter from "@/components/ProductOwnerCommandCenter.vue";

type Props = {
  usersInRoom: UserOverview;
  roundState: RoundState;
  developerDone: DeveloperDone[];
  showAllGuesses: boolean;
  hasTicketToGuess: boolean;
  hasDevelopersInRoom: boolean;
  userIsProductOwner: boolean;
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
const users = computed(() => {
  return [...props.usersInRoom.productOwnerList, ...props.usersInRoom.developerList];
});

function isDeveloper(user: (ProductOwner | Developer)): user is Developer {
  return user.role === "developer";
}

function findDeveloperDone(developer: Developer): DeveloperDone | undefined {
  return props.developerDone.find((it) => it.name === developer.name);
}

function topForElement(index: number): string {
  const theta = 2 * Math.PI * (index / users.value.length);
  const top = cy - radius * Math.cos(theta);
  return `${top}px`;
}

function leftForElement(index: number): string {
  const theta = 2 * Math.PI * (index / users.value.length);
  const left = cx + radius * Math.sin(theta);
  return `${left}px`;
}
</script>

<template>
  <div>
    <div class="virtual-table">
      <div class="table">
        <product-owner-command-center
          v-if="props.userIsProductOwner"
          :round-state="props.roundState"
          :developer-list="props.usersInRoom.developerList"
          :has-ticket-to-guess="props.hasTicketToGuess"
          :has-developers-in-room="props.hasDevelopersInRoom"
          :show-all-guesses="props.showAllGuesses"
          @estimate="emit('estimate', $event)"
          @reveal="emit('reveal')"
          @new-round="emit('new-round')"
        />
        <span v-if="!props.hasTicketToGuess && !props.userIsProductOwner">Warten auf Ticket…</span>
      </div>
      <div
        v-for="(user, index) in users"
        :key="user.name"
        class="seat"
        :style="`left:${leftForElement(index)};top:${topForElement(index)}`"
      >
        <developer-card
          v-if="isDeveloper(user)"
          :developer="user"
          :developer-done="findDeveloperDone(user)"
        />
        <span v-else>{{ user.name }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.virtual-table {
  position: relative;
  width: 37.5rem;
  height: 37.5rem;
}

.seat {
  position: absolute;
}

.table {
  position: absolute;
  top: 25%;
  left: 25%;
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

.table button {
  border: none;
  border-radius: 5px;
  font-size: 1rem;
  width: 9rem;
  height: 3rem;
  color: white;
}
</style>
