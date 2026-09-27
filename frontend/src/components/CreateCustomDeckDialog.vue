<script setup lang="ts">
import {ref} from "vue";

const deckName = ref("My custom deck");

const emits = defineEmits<{
  (e: "close"): void
}>();

</script>

<template>
  <v-dialog
    :model-value="true"
    width="auto"
    @update:model-value="emits('close')"
  >
    <v-card
      title="Create a custom deck"
      max-width="1000"
    >
      <v-text-field
        v-model="deckName"
        label="Deck name"
      />

      <v-text-field label="Deck values" />

      <v-card-actions>
        <v-btn />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>

</style>