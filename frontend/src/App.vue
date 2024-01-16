<script setup lang="ts">
import { RouterView } from "vue-router";
import { useWebsocketStore } from "@/stores/websocket";
import { onBeforeUnmount } from "vue";
const websocketStore = useWebsocketStore();
onBeforeUnmount(() => {
  websocketStore.disconnect();
});
</script>

<template>
  <header>
    <div class="wrapper">
      <h1>Estimation Poker</h1>
      <!--      <nav>-->
      <!--        <RouterLink to="/">Home</RouterLink>-->
      <!--      </nav>-->
    </div>
  </header>

  <RouterView />
</template>

<style scoped>
header {
  line-height: 1.5;
  max-height: 100vh;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }
}
</style>
