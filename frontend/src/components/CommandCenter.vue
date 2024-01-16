<script setup lang="ts">
import type { Role } from "@/components/types";
import { computed, type Ref, ref } from "vue";

type Props = {
  userRole: Role;
};

const props = defineProps<Props>();
const possibleCards = [
  { value: 1, subtitle: "Bis zu 4 Std." },
  { value: 2, subtitle: "Bis zu 8 Std." },
  { value: 3, subtitle: "Bis zu 3 Tagen" },
  { value: 4, subtitle: "Bis zu 5 Tagen" },
  { value: 5, subtitle: "Mehr als 5 Tage" },
];
const chosenCard: Ref<number | null | undefined> = ref(null);
const canGuess = computed(() => chosenCard.value !== null && chosenCard.value !== undefined);

function guess() {
  if (!canGuess.value) {
    return;
  }

  // send
}
</script>

<template>
  <v-item-group v-model="chosenCard" selected-class="bg-indigo-darken-2">
    <v-container>
      <v-row>
        <v-col v-for="card in possibleCards" :key="card.subtitle">
          <v-item :value="card.value" v-slot="{ selectedClass, toggle }">
            <v-card
              :class="['text-center', selectedClass]"
              variant="outlined"
              height="300"
              :link="true"
              @click="toggle"
            >
              <div class="mt-15">
                <v-card-title>{{ card.value }}</v-card-title>
                <v-card-subtitle>{{ card.subtitle }}</v-card-subtitle>
              </div>
            </v-card>
          </v-item>
        </v-col>
      </v-row>
    </v-container>
  </v-item-group>
  <v-container class="mt-5">
    <v-btn
      width="100%"
      prepend-icon="mdi-send"
      append-icon="mdi-send"
      :disabled="!canGuess"
      @click="guess"
      >Ab gehts</v-btn
    >
  </v-container>
</template>

<style scoped></style>
