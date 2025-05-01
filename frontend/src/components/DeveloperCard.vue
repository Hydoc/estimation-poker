<script setup lang="ts">
import type { Developer, DeveloperDone } from "@/components/types.ts";

type Props = {
  developer: Developer;
  developerDone?: DeveloperDone;
};
const props = withDefaults(defineProps<Props>(), {
  developerDone: undefined,
});

</script>

<template>
  <div>
    <div class="flip-card">
      <div :class="{'flip-card__inner': true, 'reveal': props.developerDone}">
        <div :class="{'flip-card__front': true, 'waiting-for-guess': !props.developerDone, 'guessed': props.developer.isDone }" />
        <div class="flip-card__back">
          <span v-if="props.developerDone">
            <v-icon v-if="props.developerDone.doSkip">mdi-coffee</v-icon>
            <strong v-else>{{ developerDone.guess }}</strong>
          </span>
        </div>
      </div>
    </div>
    <span>{{ props.developer.name }}</span>
  </div>
</template>

<style scoped>
.flip-card {
  margin: 0 auto;
  width: 2rem;
  height: 3rem;
  text-align: center;
}

.flip-card__inner {
  position: relative;
  width: 100%;
  height: 100%;
  text-align: center;
  transition: transform 0.5s;
  transform-style: preserve-3d;
  border-radius: 5px;
}

.reveal {
  transform: rotateY(180deg);
}

.flip-card__front {
  background-color: gray;
}

.flip-card__front,
.flip-card__back {
  position: absolute;
  width: 100%;
  height: 100%;
  backface-visibility: hidden;
  border-radius: 5px;
}

.flip-card__back {
  display: flex;
  justify-content: center;
  align-items: center;
  transform: rotateY(180deg);
  border: 2px solid #2196F3;
  background-color: white;
}

.waiting-for-guess {
  background-color: gray;
}

.guessed {
  background-color: #2196F3;
}

</style>