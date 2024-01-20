<script setup lang="ts">
import { computed, type Ref, ref } from "vue";
import type { VForm } from "vuetify/components";

type Props = {
  roundIsWaiting: boolean;
  hasTicketToGuess: boolean;
  hasDevelopersInRoom: boolean;
};

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
}>();

const ticketToGuess = ref("");
const form: Ref<VForm | undefined> = ref();

const ticketRules = [
  (value: string) => !!value || "Fehler: Hier müsste eigentlich was stehen",
  (value: string) => /^[A-Z]{2}-\d+$/.test(value) || "Fehler: Muss im Format ^[A-Z]{2}-\\d+$ sein",
];
const canEstimate = computed(() => ticketToGuess.value !== "" && form.value?.isValid);

function doLetEstimate() {
  if (!canEstimate.value) {
    return;
  }
  emit("estimate", ticketToGuess.value);
}
</script>

<template>
  <v-container>
    <v-form
      ref="form"
      :fast-fail="true"
      v-if="props.roundIsWaiting && props.hasDevelopersInRoom && !props.hasTicketToGuess"
      @submit.prevent="doLetEstimate"
    >
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
    <p v-else-if="!props.hasDevelopersInRoom">Warten auf Entwickler...</p>
  </v-container>
</template>

<style scoped></style>
