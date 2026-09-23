<script setup lang="ts">
import { computed, type Ref, ref } from "vue";
import type { VForm } from "vuetify/components";
import { isJust, type Maybe } from "@kaumlaut/pure/maybe";
import {type Developer, type Issue, RoundState} from "@/types/room.ts";

type Props = {
  roundState: RoundState;
  issueToGuess: Maybe<Issue>;
  showAllGuesses: boolean;
  developerList: Developer[];
};

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "reveal"): void;
  (e: "new-round"): void;
}>();

const roundCanBeRevealed = computed(() => props.roundState === RoundState.End);

const hasDevelopersInRoom = computed(() => props.developerList.length > 0);

const percentageDone = computed(() => {
  const devsThatAreDone = props.developerList.filter((dev) => dev.isDone).length;
  const totalDevs = props.developerList.length;
  return Math.round((devsThatAreDone / totalDevs) * 100);
});

</script>

<template>
  <v-container fluid>
    <div class="text-center">
      <v-progress-circular
        v-if="isJust(props.issueToGuess) && !props.showAllGuesses"
        v-model="percentageDone"
        class=""
        rotate="360"
        width="10"
        size="200"
        color="teal"
      >
        <template #default>
          <v-btn
            color="teal"
            :disabled="!roundCanBeRevealed"
            @click="emit('reveal')"
          >
            Reveal
          </v-btn>
        </template>
      </v-progress-circular>
      <v-btn
        v-if="props.showAllGuesses"
        width="100%"
        color="blue-grey"
        @click="emit('new-round')"
      >
        New round
      </v-btn>
      <p
        v-else-if="!hasDevelopersInRoom"
        class="text-center"
      >
        Waiting for developers...
      </p>
    </div>
  </v-container>
</template>

<style scoped></style>
