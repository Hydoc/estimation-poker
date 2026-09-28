import { computed, type ComputedRef, type Ref, ref } from "vue";
import { defineStore } from "pinia";

export type Deck = {
    name: string;
    value: string;
};

type UseCustomDecks = {
    customDecks: Readonly<Ref<Deck[]>>;
    addCustomDeck: (deck: Deck) => void;
};

export const useCustomDeckStore = defineStore("customDeck", (): UseCustomDecks => {
  const customDecks = ref<Deck[]>([]);

  function addCustomDeck(deck: Deck) {
    customDecks.value.push(deck);
  }

  return {
    customDecks,
    addCustomDeck,
  };
}, {
  persist: true,
});
