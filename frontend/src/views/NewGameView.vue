<script setup lang="ts">
import { computed, ref } from "vue";
import CreateCustomDeckDialog from "@/components/CreateCustomDeckDialog.vue";
import { type Deck, useCustomDeckStore } from "@/stores/customDeck.ts";

const customDeckStore = useCustomDeckStore();
const fibonacciValue = "0,1,2,3,4,5,8,13";
const showCreateCustomDeckDialog = ref(false);
const gameName = ref("");
const displayName = ref("");
const deck = ref(fibonacciValue);
const formValid = ref();

const requiredRule = (value: string) => !!value || "Is required";
const maxCharRule = (value: string) => value.length <= 20 || "Must be 20 chars or lower";

const availableDecks = computed(() => [
  {
    title: "Fibonacci (0,1,2,3,4,5,8,13)",
    value: fibonacciValue,
    action: () => (deck.value = fibonacciValue),
  },
  {
    title: "T-shirts (XS, S, M, L, XL)",
    value: "XS,S,M,L,XL",
    action: () => (deck.value = "XS,S,M,L,XL"),
  },
  ...customDeckStore.customDecks.map((it) => ({
    title: `${it.name} (${it.value.split(",").join(", ")})`,
    value: it.value,
    action: () => (deck.value = it.value),
  })),
  {
    title: "Create a custom deck...",
    action: () => {
      deck.value = fibonacciValue;
      showCreateCustomDeckDialog.value = true;
    },
    customClass: "text-blue",
  },
]);

function saveCustomDeck(deckToCreate: Deck) {
  showCreateCustomDeckDialog.value = false;
  customDeckStore.addCustomDeck(deckToCreate);
  deck.value = deckToCreate.value;
}

function start() {
  console.log(deck.value);
  console.log(gameName.value);
  console.log(displayName.value);
}
</script>

<template>
  <main>
    <create-custom-deck-dialog
      v-if="showCreateCustomDeckDialog"
      @close="showCreateCustomDeckDialog = false"
      @save="saveCustomDeck"
    />

    <v-form
      v-model="formValid"
      class="d-flex flex-column w-50 mx-auto"
      @submit.prevent="start"
    >
      <v-text-field
        v-model="gameName"
        :rules="[requiredRule, maxCharRule]"
        label="Game name"
      />
      <v-text-field
        v-model="displayName"
        :rules="[requiredRule, maxCharRule]"
        label="Your display name"
      />
      <v-select
        v-model="deck"
        label="Deck"
        :items="availableDecks"
      >
        <template #item="{ props, item }">
          <v-list-item
            :class="item.customClass"
            v-bind="props"
            @click="item.action"
          />
        </template>
      </v-select>

      <v-btn
        color="primary"
        :disabled="!formValid"
        type="submit"
      >
        Start
      </v-btn>

      <pre>
        {
          gameName: {{ gameName }},
          displayName: {{ displayName }},
          deck: {{ deck }},
        }
      </pre>
    </v-form>
  </main>
</template>

<style scoped></style>