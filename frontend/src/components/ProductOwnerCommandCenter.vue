<script setup lang="ts">
import { computed, ref } from "vue";

const ticketToGuess = ref("");

const ticketRules = [
  (value: string) => !!value || "Fehler: Hier müsste eigentlich was stehen",
  (value: string) => /^[A-Z]{2}-\d+$/.test(value) || "Fehler: Muss im Format ^[A-Z]{2}-\\d+$ sein",
];
const canEstimate = computed(() => ticketToGuess.value !== "");

function doLetEstimate() {}
</script>

<template>
  <v-container>
    <v-form :fast-fail="true" @submit.prevent="doLetEstimate">
      <v-row>
        <v-col>
          <v-text-field
            label="Ticket zum schätzen"
            :rules="ticketRules"
            v-model="ticketToGuess"
            placeholder="CC-0000"
            required
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col class="text-right">
          <v-btn type="submit" :disabled="!canEstimate">Schätzen lassen</v-btn>
        </v-col>
      </v-row>
    </v-form>
  </v-container>
</template>

<style scoped></style>
