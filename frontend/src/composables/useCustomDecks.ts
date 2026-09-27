import {computed, type ComputedRef, ref} from "vue";

type Deck = {
    name: string;
    value: string;
};

type State = {
    decks: Readonly<Deck[]>;
};

type UseCustomDecks = {
    state: Readonly<ComputedRef<State>>;
};

export function useCustomDecks(): UseCustomDecks {
    const customDecks = ref<Deck[]>([{ name: "Gutes Deck", value: "1,2,3,4,5" }]);
    
    const state = computed((): State => ({
        decks: customDecks.value,
    }));
    
    function add(deck: Deck) {
        customDecks.value.push(deck);
    }
    
    return {
        state,
    };
}