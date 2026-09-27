<script setup lang="ts">
import {useRouter} from "vue-router";

const router = useRouter();

function startNewGame() {
  router.push({
    name: "newGame",
  });
}
</script>

<template>
  <main>
    <div class="d-flex flex-column align-center ga-5">
      <h1 class="text-headline-large d-flex flex-column align-center">
        <span>Estimation Poker</span>
        <span>Play. Estimate. Deliver.</span>
      </h1>
      <v-btn
        size="x-large"
        color="primary"
        @click="startNewGame"
      >
        Start a new game!
      </v-btn>
    </div>
  </main>
</template>

<style scoped>

</style>