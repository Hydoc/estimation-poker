import type {Ref} from "vue";
import {ref} from "vue";
import {defineStore} from "pinia";
import {type PossibleGuess,} from "@/components/types";

type WebsocketStore = {
  userExistsInRoom(name: string, roomId: string): Promise<boolean>;
  fetchPossibleGuesses(): Promise<void>;
  possibleGuesses: Ref<PossibleGuess[]>;
};

export const useWebsocketStore = defineStore("websocket", (): WebsocketStore => {
  const possibleGuesses: Ref<PossibleGuess[]> = ref([]);

  async function userExistsInRoom(name: string, roomId: string): Promise<boolean> {
    const response = await fetch(`/v1/room/${roomId}/users/exists?name=${name}`);
    return ((await response.json()) as { exists: boolean }).exists;
  }

  async function fetchPossibleGuesses() {
    const response = await fetch("/v1/possible-guesses");
    if (!response.ok) {
      possibleGuesses.value = [];
      return;
    }

    possibleGuesses.value = await response.json();
  }

  return {
    possibleGuesses,
    userExistsInRoom,
    fetchPossibleGuesses,
  };
});
