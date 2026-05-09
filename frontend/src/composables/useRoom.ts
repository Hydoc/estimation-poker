import { computed, type ComputedRef, type Ref, ref } from "vue";
import { isJust, just, type Maybe, nothing } from "@kaumlaut/pure/maybe";
import {
  isPermissions,
  isRoomStateResponse,
  isUserOverview,
  type ReceivableWebsocketMessage,
  type RoomState,
  type Round, type SendableWebsocketMessage,
} from "@/types/room.ts";
import { useWebsocket } from "@/composables/useWebsocket.ts";
import {
  type DeveloperDone,
  type Permissions,
  Role,
  RoundState,
  type UserOverview,
} from "@/components/types.ts";
import {
  attemptErrorAware,
  fail,
  type FetchState,
  isSuccess,
  none,
} from "@kaumlaut/pure/fetch-state";

export type UseRoom = {
  roomState: ComputedRef<RoomState>;
  joinRoom(name: string, role: Role, roomId: string): Promise<void>;
  leaveRoom(): void;
  send(message: SendableWebsocketMessage): void;
  fetchRoomState(roomId: string): Promise<{ isLocked: boolean; inProgress: boolean }>;
  fetchPermissions(): void;
};

export function useRoom(): UseRoom {
  const websocket = useWebsocket();
  const roomId = ref<Maybe<string>>(nothing());
  const name = ref<Maybe<string>>(nothing());
  const role = ref<Maybe<Role>>(nothing());
  const guess = ref<Maybe<number>>(nothing());
  const issueToGuess = ref<Maybe<string>>(nothing());
  const doSkip = ref<boolean>(false);
  const roundState = ref<RoundState>(RoundState.Waiting);
  const users = ref<FetchState<UserOverview>>(none());
  const notifications = ref<string[]>([]);
  const showAllGuesses = ref<boolean>(false);
  const roomIsLocked = ref<boolean>(false);
  const roundInProgress = ref<boolean>(false);
  const developerDone: Ref<DeveloperDone[]> = ref([]);
  const issues = ref<any[]>([]);
  const permissions = ref<Permissions>({
    room: {
      canLock: false,
    },
  });

  const roomState = computed(
    (): RoomState => ({
      id: roomId.value,
      guess: guess.value,
      role: role.value,
      name: name.value,
      doSkip: doSkip.value,
      issueToGuess: issueToGuess.value,
      roundState: roundState.value,
      users: users.value,
      notifications: notifications.value,
      showAllGuesses: showAllGuesses.value,
      roomIsLocked: roomIsLocked.value,
      roundInProgress: roundInProgress.value,
      developerDone: developerDone.value,
      issues: issues.value,
      isConnected: websocket.isConnected.value,
      permissions: permissions.value,
    }),
  );

  const roomStateAsRound = computed((): Round => {
    if (!isJust(issueToGuess.value) || !isSuccess(users.value)) {
      throw new Error("round is invalid");
    }

    return {
      developerDone: developerDone.value,
      issueToGuess: issueToGuess.value.value,
      state: roundState.value,
      users: users.value.data,
    };
  });

  function resetRound() {
    issueToGuess.value = nothing();
    guess.value = nothing();
    doSkip.value = false;
    roundState.value = RoundState.Waiting;
    showAllGuesses.value = false;
    developerDone.value = [];
  }

  async function joinRoom(username: string, userRole: Role, roomIdToJoin: string) {
    const roleUrl = userRole === Role.Developer ? "developer" : "product-owner";
    const url = `${window.location.host}/v1/room/${roomIdToJoin}/${roleUrl}?name=${username}`;
    const connected = await websocket.connect(url, onWebsocketMessage);
    if (!connected) {
      throw new Error("could not connect");
    }

    roomId.value = just(roomIdToJoin);
    role.value = just(userRole);
    name.value = just(username);
  }
  
  function send(message: SendableWebsocketMessage) {
    websocket.send(message);
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
        issueToGuess.value = just(decoded.data);
        break;
      case "you-guessed":
        guess.value = just(decoded.data);
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
      users.value = none();
      return;
    }
    const response = await fetch(`/v1/room/${roomId.value.value}/users`);
    if (!response.ok) {
      users.value = fail("error trying to fetch users");
      return;
    }

    users.value = attemptErrorAware(isUserOverview)(await response.json());
  }

  function leaveRoom() {
    websocket.disconnect();
    roomId.value = nothing();
    notifications.value = [];
    name.value = nothing();
    role.value = nothing();
    permissions.value = {
      room: {
        canLock: false,
      },
    };
    resetRound();
  }

  async function fetchRoomState(
    roomId: string,
  ): Promise<{ isLocked: boolean; inProgress: boolean }> {
    const response = await fetch(`/v1/room/${roomId}/state`);

    if (!response.ok) {
      throw new Error("Could not fetch room state");
    }

    const result = isRoomStateResponse(await response.json());
    if (!result.success) {
      console.log(result.errors);
      throw new Error("Room state is invalid");
    }

    return result.value;
  }

  async function fetchPermissions() {
    if (!isJust(roomId.value) || !isJust(name.value)) {
      throw new Error("could not fetch permissions");
    }

    const response = await fetch(
      `/v1/room/${roomId.value.value}/permissions?name=${name.value.value}`,
    );
    if (!response.ok) {
      permissions.value = {
        room: {
          canLock: false,
        },
      };
      return;
    }

    const result = isPermissions(await response.json());
    if (!result.success) {
      console.log(result.errors);
      throw new Error("permissions is not valid");
    }

    permissions.value = result.value.permissions;
  }

  return {
    roomState,
    joinRoom,
    leaveRoom,
    send,
    fetchRoomState,
    fetchPermissions,
  };
}
