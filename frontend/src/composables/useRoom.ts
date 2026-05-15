import { computed, type ComputedRef, type Ref, ref } from "vue";
import { isJust, just, type Maybe, nothing } from "@kaumlaut/pure/maybe";
import {
  type ConnectionState,
  type DeveloperDone,
  isConnectionState,
  isPermissions,
  isRoomMetadata,
  isRoomStateResponse,
  type Issue,
  isUserOverview,
  type Permissions,
  type PossibleGuess,
  type ReceivableWebsocketMessage,
  Role,
  type RoomMetadata,
  type RoomState,
  RoundState,
  type SendableWebsocketMessage,
  type UserOverview,
} from "@/types/room.ts";
import { useWebsocket } from "@/composables/useWebsocket.ts";
import { attemptErrorAware, fail, type FetchState, none } from "@kaumlaut/pure/fetch-state";

export type UseRoom = {
  roomState: ComputedRef<RoomState>;
  roomNotifications: Ref<string[]>;
  joinRoom(name: string, role: Role, roomId: string): Promise<void>;
  leaveRoom(): void;
  send(message: SendableWebsocketMessage): void;
  roomMetadata(roomId: string): Promise<RoomMetadata>;
  connectionState(roomId: string, username: string, password: string): Promise<ConnectionState>;
  fetchRoomState(): Promise<void>;
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
  const roomNotifications = ref<string[]>([]);
  const showAllGuesses = ref<boolean>(false);
  const roomIsLocked = ref<boolean>(false);
  const roundInProgress = ref<boolean>(false);
  const developerDone: Ref<DeveloperDone[]> = ref([]);
  const issues = ref<Issue[]>([]);
  const possibleGuesses = ref<PossibleGuess[]>([]);
  const permissions = ref<Permissions>({
    canLockRoom: false,
    key: "",
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
      showAllGuesses: showAllGuesses.value,
      roomIsLocked: roomIsLocked.value,
      roundInProgress: roundInProgress.value,
      developerDone: developerDone.value,
      issues: issues.value,
      isConnected: websocket.isConnected.value,
      permissions: permissions.value,
      possibleGuesses: possibleGuesses.value,
    }),
  );

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
      throw new Error("Could not connect");
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
        roomNotifications.value.push(`${decoded.data} has left the room…`);
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
      case "issues":
        await fetchRoomState();
        break;
      case "permissions":
        const result = isPermissions(decoded.data);
        if (!result.success) {
          console.error(result.errors);
          throw new Error("permissions is invalid");
        }

        permissions.value = result.value;
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
    roomNotifications.value = [];
    name.value = nothing();
    role.value = nothing();
    issues.value = [];
    permissions.value = {
      canLockRoom: false,
      key: "",
    };
    resetRound();
  }

  async function roomMetadata(roomId: string): Promise<RoomMetadata> {
    const response = await fetch(`/v1/room/${roomId}/metadata`);

    if (!response.ok) {
      throw new Error("Could not fetch room metadata");
    }

    const result = isRoomMetadata(await response.json());

    if (!result.success) {
      console.error(result.errors);
      throw new Error("Room metadata is invalid");
    }

    return result.value;
  }

  async function connectionState(
    roomId: string,
    username: string,
    password: string,
  ): Promise<ConnectionState> {
    const response = await fetch(`/v1/room/${roomId}/connection-state`, {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      throw new Error("Could not fetch connection state");
    }

    const result = isConnectionState(await response.json());
    if (!result.success) {
      console.error(result.errors);
      throw new Error("Connection state is invalid");
    }

    return result.value;
  }

  async function fetchRoomState() {
    if (!isJust(roomId.value)) {
      throw new Error("Could not fetch room state");
    }
    const response = await fetch(`/v1/room/${roomId.value.value}/state`);

    if (!response.ok) {
      throw new Error("Could not fetch room state");
    }

    const result = isRoomStateResponse(await response.json());
    if (!result.success) {
      console.error(result.errors);
      throw new Error("Room state is invalid");
    }

    issues.value = result.value.issues;
    roomIsLocked.value = result.value.isLocked;
    roundInProgress.value = result.value.inProgress;
    possibleGuesses.value = result.value.possibleGuesses;
  }

  return {
    roomState,
    roomNotifications,
    joinRoom,
    leaveRoom,
    send,
    roomMetadata,
    connectionState,
    fetchRoomState,
  };
}
