import { computed, type ComputedRef, type Ref, ref } from "vue";
import { isJust, just, type Maybe, nothing } from "@kaumlaut/pure/maybe";
import {
  type ConnectionState,
  type DeveloperDone,
  isConnectionState,
  isEstimateWebsocketMessage,
  isEveryoneDoneWebsocketMessage,
  isIssuesWebsocketMessage,
  isLeaveWebsocketMessage,
  isNewRoundWebsocketMessage,
  isPermissionsWebsocketMessage,
  isReceivableWebsocketMessage,
  isRevealWebsocketMessage,
  isRoomLockedWebsocketMessage,
  isRoomMetadata,
  isRoomOpenedWebsocketMessage,
  isRoomStateResponse,
  type Issue,
  isUsersWebsocketMessage,
  isYouGuessedWebsocketMessage,
  isYouSkippedWebsocketMessage,
  type Permissions,
  type PossibleGuess,
  Role,
  type RoomMetadata,
  type RoomState,
  RoundState,
  type SendableWebsocketMessage,
  type UserOverview,
} from "@/types/room.ts";
import { useWebsocket } from "@/composables/useWebsocket.ts";

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
  const users = ref<Maybe<UserOverview>>(nothing());
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
    const result = isReceivableWebsocketMessage(JSON.parse(message.data));
    if (!result.success) {
      throw new Error(`message is invalid`);
    }

    if (isLeaveWebsocketMessage(result.value).success) {
      roomNotifications.value.push(`${result.value.data} has left the room…`);
      return;
    }

    if (isUsersWebsocketMessage(result.value).success) {
      users.value = just(result.value.data);
      return;
    }

    if (isEstimateWebsocketMessage(result.value).success) {
      roundState.value = RoundState.InProgress;
      issueToGuess.value = just(result.value.data);
      return;
    }

    if (isYouGuessedWebsocketMessage(result.value).success) {
      guess.value = just(result.value.data);
      doSkip.value = false;
      return;
    }

    if (isYouSkippedWebsocketMessage(result.value).success) {
      guess.value = nothing();
      doSkip.value = true;
      return;
    }

    if (isEveryoneDoneWebsocketMessage(result.value).success) {
      roundState.value = RoundState.End;
      return;
    }

    if (isRevealWebsocketMessage(result.value).success) {
      developerDone.value = result.value.data;
      showAllGuesses.value = true;
      return;
    }

    if (isRoomLockedWebsocketMessage(result.value).success) {
      roomIsLocked.value = true;
      return;
    }

    if (isRoomOpenedWebsocketMessage(result.value).success) {
      roomIsLocked.value = false;
      return;
    }

    if (isNewRoundWebsocketMessage(result.value).success) {
      resetRound();
      return;
    }

    if (isIssuesWebsocketMessage(result.value).success) {
      await fetchRoomState();
      return;
    }

    if (isPermissionsWebsocketMessage(result.value).success) {
      permissions.value = result.value.data;
      return;
    }

    throw new Error(`websocket message ${result.value.type} is not registered.`);
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
