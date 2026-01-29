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
          v-for="possibleGuess in props.possibleGuesses"
          :key="possibleGuess.guess"
          :class="{
            card: true,
            'active-guess': props.guess === possibleGuess.guess && !props.didSkip,
          }"
          @click="doGuess(possibleGuess.guess)"
        >
          <h2>{{ possibleGuess.guess }}</h2>
          <span class="guess-description">{{ possibleGuess.description }}</span>
        </div>
        <div
          :class="{ card: true, 'active-guess': props.didSkip }"
          @click="skip"
        >
          <h2><v-icon>mdi-coffee</v-icon></h2>
          <span class="guess-description">Skip round</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card {
  padding: 0.5rem;
  z-index: 1;
  background-color: white;
  box-shadow: 0 1px 2px 1px rgba(0, 0, 0, 0.4);
  user-select: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  transition: 0.2s;
  min-height: 11rem;
  min-width: 7rem;
  max-width: 7rem;
}

.card:not(:first-child) {
  margin-left: calc(2rem * -1);
}

.active-guess,
.card:hover ~ .card,
.card:focus-within ~ .card {
  transform: translateX(2rem);
}

.card:hover,
.card:focus-within,
.active-guess {
  transform: translateY(-1rem);
  cursor: pointer;
  background-color: #f0f8ff;
}

.guess-description {
  text-align: center;
  opacity: 0.6;
}
</style>
