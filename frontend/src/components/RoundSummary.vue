<script setup lang="ts">
import type { DeveloperDone } from "@/components/types.ts";
import { computed } from "vue";

type Props = {
  developerDone: DeveloperDone[];
};
const props = defineProps<Props>();

const totalDevelopers = computed(() => props.developerDone.length);
const stats = computed(() => {
  return props.developerDone.reduce((carry: { [key: number]: number }, it) => {
    if (!carry[it.guess]) {
      carry[it.guess] = 0;
    }
    carry[it.guess]++;
    return carry;
  }, {});
});

const mostGuessedPercentage = computed(() => {
  const amount = Math.max(...Object.values(stats.value));
  return (amount / totalDevelopers.value) * 100;
});
</script>

<template>
  <v-bottom-sheet :model-value="true" inset height="250" :scrim="false">
    <v-card>
      <v-card-text class="d-flex ga-5 pt-16 justify-center align-center">
        <div v-for="(amountOfGuesses, stat) in stats" :key="stat">
          <div class="d-flex flex-column justify-center align-center ga-2">
            <progress class="progress" :max="totalDevelopers" :value="amountOfGuesses" />
            <div class="card">
              <span>
                <strong v-if="stat > 0">{{ stat }}</strong>
                <v-icon v-else>mdi-coffee</v-icon>
              </span>
            </div>
            <span
              ><strong
                >{{ amountOfGuesses }}
                {{ amountOfGuesses === 1 ? "Schätzung" : "Schätzungen" }}</strong
              ></span
            >
          </div>
        </div>

        <div class="d-flex flex-column justify-center align-center ga-3 pb-5 pl-4">
          <strong class="agreement">Übereinstimmung:</strong>
          <v-progress-circular
            size="50"
            color="teal"
            width="10"
            :model-value="mostGuessedPercentage"
          />
        </div>
      </v-card-text>
    </v-card>
  </v-bottom-sheet>
</template>

<style scoped>
.card {
  width: 2rem;
  height: 3rem;
  border: 2px solid #36454f;
  display: flex;
  border-radius: 5px;
  align-items: center;
  justify-content: center;
}

.agreement {
  font-size: 1.1rem;
  opacity: 0.4;
}

progress {
  width: 4rem;
  height: 0.25rem;
  transform: rotate(-90deg);
  margin-bottom: 1.75rem;
}

progress[value] {
  -webkit-appearance: none;
  appearance: none;
}

progress[value]::-webkit-progress-bar {
  background-color: #eee;
  border-radius: 5px;
}

progress[value]::-webkit-progress-value {
  background-color: #36454f;
  border-radius: 5px;
}
</style>
