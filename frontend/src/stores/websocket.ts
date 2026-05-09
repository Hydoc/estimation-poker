import type {Ref} from "vue";
import {ref} from "vue";
import {defineStore} from "pinia";
import {type PossibleGuess,} from "@/components/types";

type WebsocketStore = {
  userExistsInRoom(name: string, roomId: string): Promise<boolean>;
  fetchPossibleGuesses(): Promise<void>;
  passwordMatchesRoom(roomId: string, password: string): Promise<boolean>;
  possibleGuesses: Ref<PossibleGuess[]>;
};

export const useWebsocketStore = defineStore("websocket", (): WebsocketStore => {
  const possibleGuesses: Ref<PossibleGuess[]> = ref([]);

  async function userExistsInRoom(name: string, roomId: string): Promise<boolean> {
    const response = await fetch(`/v1/room/${roomId}/users/exists?name=${name}`);
    return ((await response.json()) as { exists: boolean }).exists;
  }

  async function passwordMatchesRoom(roomId: string, password: string): Promise<boolean> {
    const response = await fetch(`/v1/room/${roomId}/authenticate`, {
      method: "POST",
      body: JSON.stringify({ password }),
    });

    if (!response.ok) {
      return false;
    }

    return ((await response.json()) as { ok: boolean }).ok;
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
    passwordMatchesRoom,
  };
});
