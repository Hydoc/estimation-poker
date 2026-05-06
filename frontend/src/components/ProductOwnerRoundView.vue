<script setup lang="ts">
import { computed, type Ref, ref } from "vue";
import type { VForm } from "vuetify/components";
import { type Developer, RoundState } from "@/components/types.ts";

type Props = {
  roundState: RoundState;
  hasTicketToGuess: boolean;
  actualTicketToGuess: string;
  showAllGuesses: boolean;
  developerList: Developer[];
};

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "reveal"): void;
  (e: "new-round"): void;
}>();

const ticketToGuess = ref("");
const form: Ref<VForm | undefined> = ref();

const ticketRules = [
  (value: string) => !!value || "Error: Can not be empty",
  (value: string) => /^[A-Z]{2,}-\d+$/.test(value) || "Error: ^[A-Z]{2,}-\\d+$ required",
];
const canEstimate = computed(() => ticketToGuess.value !== "" && form.value?.isValid);

const roundIsWaiting = computed(() => props.roundState === RoundState.Waiting);

const roundCanBeRevealed = computed(() => props.roundState === RoundState.End);

const hasDevelopersInRoom = computed(() => props.developerList.length > 0);

const percentageDone = computed(() => {
  const devsThatAreDone = props.developerList.filter((dev) => dev.isDone).length;
  const totalDevs = props.developerList.length;
  return Math.round((devsThatAreDone / totalDevs) * 100);
});

function doLetEstimate() {
  if (!canEstimate.value) {
    return;
  }
  emit("estimate", ticketToGuess.value);
  ticketToGuess.value = "";
}
</script>

<template>
  <v-container fluid>
    <div class="text-center">
      <v-form
        v-if="roundIsWaiting && hasDevelopersInRoom && !props.hasTicketToGuess"
        ref="form"
        :fast-fail="true"
        @submit.prevent="doLetEstimate"
      >
        <v-text-field
          v-model="ticketToGuess"
          bg-color="white"
          label="Ticket to guess"
          :rules="ticketRules"
          placeholder="CC-0000"
          required
        />
        <v-btn width="100%" type="submit" :disabled="!canEstimate"> Estimate </v-btn>
      </v-form>
      <v-progress-circular
        v-if="props.hasTicketToGuess && !props.showAllGuesses"
        v-model="percentageDone"
        class=""
        rotate="360"
        width="10"
        size="200"
        color="teal"
      >
        <template #default>
          <v-btn color="teal" :disabled="!roundCanBeRevealed" @click="emit('reveal')">
            Reveal
          </v-btn>
        </template>
      </v-progress-circular>
      <v-btn v-if="props.showAllGuesses" width="100%" color="blue-grey" @click="emit('new-round')">
        New round
      </v-btn>
      <p v-else-if="!hasDevelopersInRoom" class="text-center">Waiting for developers...</p>
    </div>
  </v-container>
</template>

<style scoped></style>
