import {computed, type ComputedRef, type Ref, ref} from "vue";
import {isJust, just, type Maybe, nothing} from "@kaumlaut/pure/maybe";
import type {ReceivableWebsocketMessage, RoomState} from "@/types/room.ts";
import {useWebsocket} from "@/composables/useWebsocket.ts";
import {type DeveloperDone, Role, RoundState, type UserOverview} from "@/components/types.ts";

export type UseRoom = {
  roomState: ComputedRef<RoomState>;
  setRoomId: (id: Maybe<string>) => void;
  join(name: string, role: Role, roomId: string): Promise<void>;
  leave(): void;
};

export function useRoom(): UseRoom {
  const websocket = useWebsocket();
  const roomId = ref<Maybe<string>>(nothing());
  const guess = ref<Maybe<number>>(nothing());
  const issueToGuess = ref<Maybe<string>>(nothing());
  const doSkip = ref<boolean>(false);
  const roundState = ref<RoundState>(RoundState.Waiting);
  const users = ref<UserOverview>([]);
  const notifications = ref<string[]>([]);
  const showAllGuesses = ref<boolean>(false);
  const roomIsLocked = ref<boolean>(false);
  const developerDone: Ref<DeveloperDone[]> = ref([]);
  const issues = ref<any[]>([]);

  const roomState = computed((): RoomState => ({
    id: roomId.value,
    guess: guess.value,
    doSkip: doSkip.value,
    issueToGuess: issueToGuess.value,
    roundState: roundState.value,
    users: users.value,
    notifications: notifications.value,
    showAllGuesses: showAllGuesses.value,
    roomIsLocked: roomIsLocked.value,
    developerDone: developerDone.value,
    issues: issues.value,
  }));
  
  function resetRound() {
    issueToGuess.value = nothing();
    guess.value = nothing();
    doSkip.value = false;
    roundState.value = RoundState.Waiting;
    showAllGuesses.value = false;
    developerDone.value = [];
  }

  function setRoomId(id: Maybe<string>) {
    roomId.value = id;
  }
  
  async function join(name: string, role: Role, roomId: string) {
    const roleUrl = role === Role.Developer ? "developer" : "product-owner";
    const url = `${window.location.host}/v1/room/${roomId}/${roleUrl}?name=${name}`;
    const connected = await websocket.connect(url, onWebsocketMessage);
    if (!connected) {
      throw new Error("could not connect");
    }
  }
  
  async function onWebsocketMessage(message: MessageEvent): Promise<void> {
    const decoded = JSON.parse(message.data) as ReceivableWebsocketMessage;
    switch (decoded.type) {
      case "leave":
        await fetchUsersInRoom();
        notifications.value.push(`${decoded.data} has left the room…`);
        break;
      case "join":
      case "developer-skipped":
      case "developer-guessed":
        await fetchUsersInRoom();
        break;
      case "estimate":
        roundState.value = RoundState.InProgress;
        issueToGuess.value = decoded.data;
        break;
      case "you-guessed":
        guess.value = decoded.data;
        doSkip.value = false;
        break;
      case "you-skipped":
        guess.value = nothing();
        doSkip.value = true;
        break;
      case "everyone-done":
        await fetchUsersInRoom();
        roundState.value = RoundState.End;
        break;
      case "reveal":
        developerDone.value = decoded.data;
        showAllGuesses.value = true;
        break;
      case "room-locked":
        roomIsLocked.value = true;
        break;
      case "room-opened":
        roomIsLocked.value = false;
        break;
      case "new-round":
        resetRound();
        await fetchUsersInRoom();
        break;
    }
  }

  async function fetchUsersInRoom() {
    if (!isJust(roomId.value)) {
      users.value = [];
      return;
    }
    const response = await fetch(`/v1/room/${roomId.value.value}/users`);
    if (!response.ok) {
      users.value = [];
      return;
    }

    users.value = await response.json();
  }
  
  function leave() {
    websocket.disconnect();
    notifications.value = [];
    resetRound();
  }

  // async function fetchRoomState() {
  //     state.value = load();
  //
  //     const response = await fetch(`/v1/room/${roomId}/state`);
  //
  //     state.value = attemptErrorAware(isRoomState)(d)
  // }

  return {
    roomState,
    setRoomId,
    join,
    leave,
  };
}
