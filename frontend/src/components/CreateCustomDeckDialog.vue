<script setup lang="ts">
import { computed, ref } from "vue";
import FlipCard from "@/components/FlipCard.vue";
import type { Deck } from "@/stores/customDeck.ts";

const defaultDeckName = "My custom deck";
const defaultDeckValues = "1,2,3";

const deckName = ref(defaultDeckName);
const deckValues = ref(defaultDeckValues);
const formValid = ref();

const deckValuesAsList = computed(() => deckValues.value.split(","));

const deckValueRules = [
  (value: string) => {
    if (!isDeckValueValid(value)) {
      return "Three numbers per card, comma separated, at least 2 cards"
    }

    if (!isDeckUnique(value)) {
      return "Cards must be unique";
    }

    if (!isDeckCorrectSize(value)) {
      return "Maximum of 10 cards";
    }
    return true;
  }
];

function isDeckValueValid(value: string): boolean {
  return /^[^,].{1,3}(,[^,]{1,3})+$/.test(value);
}

function isDeckUnique(value: string): boolean {
  return [...new Set(value.split(","))].length === value.split(",").length
}

function isDeckCorrectSize(value: string): boolean {
  return value.split(",").length <= 10;
}

function save() {
  emits("save", { name: deckName.value, value: deckValues.value });
  deckName.value = defaultDeckName;
  deckValues.value = defaultDeckValues;
}

const emits = defineEmits<{
  (e: "close"): void;
  (e: "save", deck: Deck): void;
}>();
</script>

<template>
  <v-dialog
    :model-value="true"
    max-width="1000"
    @update:model-value="emits('close')"
  >
    <v-card title="Create a custom deck">
      <v-form
        v-model="formValid"
        @submit.prevent="save"
      >
        <v-card-text>
          <v-text-field
            v-model="deckName"
            label="Deck name"
          />

          <v-text-field
            v-model="deckValues"
            label="Deck values"
            hint="Three characters per card, comma separated, maximum of 10 unique cards, at least 2 cards"
            :rules="deckValueRules"
            persistent-hint
            @keydown.space.prevent
          />

          <div>
            <h3>Preview</h3>
            <p>This is a preview of how your deck will look</p>

            <div class="d-flex ga-1 w-100 pa-2">
              <flip-card
                v-for="card in deckValuesAsList"
                :key="card"
                :value="card"
                :reveal="true"
                :estimated="true"
              />
            </div>
          </div>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            color="red"
            @click="emits('close')"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            type="submit"
            :disabled="!formValid"
          >
            Save deck
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </v-dialog>
</template>

<style scoped></style>