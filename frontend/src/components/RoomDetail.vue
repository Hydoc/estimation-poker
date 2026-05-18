<script setup lang="ts">
import { computed, ref, watch } from "vue";
import TableOverview from "@/components/TableOverview.vue";
import DeveloperRoundView from "@/components/DeveloperRoundView.vue";
import RoundSummary from "@/components/RoundSummary.vue";
import { Role, type RoomState } from "@/types/room.ts";
import { isJust } from "@kaumlaut/pure/maybe";
import { isSuccess } from "@kaumlaut/pure/fetch-state";

type Props = {
  roomState: RoomState;
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

const userIsDeveloper = computed(
  () => isJust(props.roomState.role) && props.roomState.role.value === Role.Developer,
);
const hasTicketToGuess = computed(() => isJust(props.roomState.issueToGuess));

watch(
  () => props.roomState.showAllGuesses,
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
    :developer-done="props.roomState.developerDone"
  />

  <v-container fluid>
    <v-col cols="12">
      <v-row>
        <table-overview
          v-if="isJust(props.roomState.users)"
          :show-all-guesses="props.roomState.showAllGuesses"
          :users-in-room="props.roomState.users.value"
          :round-state="props.roomState.roundState"
          :user-role="props.roomState.role"
          :developer-done="props.roomState.developerDone"
          :issue-to-guess="props.roomState.issueToGuess"
          @estimate="emit('estimate', $event)"
          @reveal="emit('reveal')"
          @new-round="emit('new-round')"
        />
      </v-row>

      <v-row
        align="center"
        justify="center"
      >
        <developer-round-view
          v-if="userIsDeveloper"
          class="developer-command-center"
          :show-all-guesses="props.roomState.showAllGuesses"
          :guess="props.roomState.guess"
          :did-skip="props.roomState.doSkip"
          :has-issue-to-guess="hasTicketToGuess"
          :possible-guesses="props.roomState.possibleGuesses"
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
