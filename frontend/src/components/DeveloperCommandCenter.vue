<script setup lang="ts">
import { type PossibleGuess } from "@/components/types";

type Props = {
  guess: number;
  didSkip: boolean;
  showAllGuesses: boolean;
  hasTicketToGuess: boolean;
  possibleGuesses: PossibleGuess[];
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "guess", guess: number): void;
  (e: "skip"): void;
}>();

function doGuess(value: number) {
  emit("guess", value);
}

function skip() {
  emit("skip");
}
</script>

<template>
  <div class="d-flex flex-column justify-center align-center">
    <div v-if="props.hasTicketToGuess && !props.showAllGuesses">
      <div class="d-flex ga-2">
        <div
          :class="{
            card: true,
            'active-guess': props.guess === possibleGuess.guess && !props.didSkip,
          }"
          v-for="possibleGuess in props.possibleGuesses"
          :key="possibleGuess.guess"
          @click="doGuess(possibleGuess.guess)"
        >
          <h2>{{ possibleGuess.guess }}</h2>
          <span class="guess-description">{{ possibleGuess.description }}</span>
        </div>
        <v-btn
          class="align-self-center ml-10"
          @click="skip"
          :icon="props.didSkip ? `mdi-coffee-outline` : `mdi-coffee`"
          title="Runde aussetzen"
          :color="props.didSkip ? `#38220f` : `#967259`"
        ></v-btn>
      </div>
    </div>
    <p v-if="!props.hasTicketToGuess">Warten auf Ticket...</p>
  </div>
</template>

<style scoped>
.card {
  display: flex;
  flex-direction: column;
  border: 1px solid rgba(0, 0, 0, 0.5);
  align-items: center;
  justify-content: center;
  border-radius: 5%;
  transition: 0.4s;
  min-height: 15rem;
  min-width: 10rem;
}
.card:hover,
.active-guess {
  cursor: pointer;
  background-color: #82b1ff;
  transform: translate(0, -10px);
}
.guess-description {
  opacity: 0.6;
}
</style>
