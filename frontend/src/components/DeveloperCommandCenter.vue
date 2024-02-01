<script setup lang="ts">
import { computed, ref, type Ref } from "vue";
import { PossibleGuess } from "@/components/types";

type Props = {
  didGuess: boolean;
  hasTicketToGuess: boolean;
  possibleGuesses: PossibleGuess[];
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "guess", guess: number): void;
}>();

const chosenCard: Ref<number | null | undefined> = ref(null);
const canGuess = computed(() => chosenCard.value !== null && chosenCard.value !== undefined);

function guess() {
  if (!canGuess.value) {
    return;
  }

  emit("guess", chosenCard.value!);
  chosenCard.value = null;
}
</script>

<template>
  <v-container>
    <v-item-group
      v-if="props.hasTicketToGuess && !props.didGuess"
      v-model="chosenCard"
      selected-class="bg-indigo-darken-2"
    >
      <v-container>
        <v-row>
          <v-col v-for="possibleGuess in props.possibleGuesses" :key="possibleGuess.guess">
            <v-item :value="possibleGuess.guess" v-slot="{ selectedClass, toggle }">
              <v-card
                :class="['text-center', selectedClass]"
                variant="outlined"
                height="300"
                :link="true"
                @click="toggle"
              >
                <div class="mt-15">
                  <v-card-title>{{ possibleGuess.guess }}</v-card-title>
                  <v-card-subtitle>{{ possibleGuess.description }}</v-card-subtitle>
                </div>
              </v-card>
            </v-item>
          </v-col>
        </v-row>
      </v-container>
    </v-item-group>
    <v-btn
      v-if="props.hasTicketToGuess && !props.didGuess"
      width="100%"
      prepend-icon="mdi-send"
      append-icon="mdi-send"
      :disabled="!canGuess"
      @click="guess"
      >Ab gehts</v-btn
    >
    <p v-else-if="!props.hasTicketToGuess">Warten auf Ticket...</p>
  </v-container>
</template>

<style scoped></style>
