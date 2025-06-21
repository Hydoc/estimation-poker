import type { Ref } from "vue";
import { computed, ref } from "vue";
import { defineStore } from "pinia";
import {
  type PossibleGuess,
  type UserOverview,
  type Permissions,
  type DeveloperDone,
} from "@/components/types";
import { Role, RoundState } from "@/components/types";

type WebsocketStore = {
  connect(name: string, role: string, roomId: string): Promise<void>;
  disconnect(): void;
  resetRound(): void;
  userExistsInRoom(name: string, roomId: string): Promise<boolean>;
  fetchActiveRooms(): Promise<string[]>;
  send(message: SendableWebsocketMessage): void;
  isRoundInRoomInProgress(roomId: string): Promise<boolean>;
  isRoomLocked(roomId: string): Promise<boolean>;
  fetchPossibleGuesses(): Promise<void>;
  fetchPermissions(): Promise<void>;
  fetchRoomIsLocked(): Promise<boolean>;
  passwordMatchesRoom(roomId: string, password: string): Promise<boolean>;
  username: Ref<string>;
  isConnected: Ref<boolean>;
  usersInRoom: Ref<UserOverview>;
  roomId: Ref<string>;
  userRole: Ref<Role>;
  roundState: Ref<RoundState>;
  ticketToGuess: Ref<string>;
  guess: Ref<number>;
  didSkip: Ref<boolean>;
  showAllGuesses: Ref<boolean>;
  possibleGuesses: Ref<PossibleGuess[]>;
  permissions: Ref<Permissions>;
  roomIsLocked: Ref<boolean>;
  developerDone: Ref<DeveloperDone[]>;
};

export type SendableWebsocketMessageType =
  | "estimate"
  | "guess"
  | "reveal"
  | "new-round"
  | "lock-room"
  | "skip"
  | "open-room";

type SendableWebsocketMessage = {
  type: SendableWebsocketMessageType;
  data?: any;
};

type ReceivableWebsocketMessage = {
  type:
    | "join"
    | "leave"
    | "estimate"
    | "reveal"
    | "developer-guessed"
    | "everyone-done"
    | "you-guessed"
    | "you-skipped"
    | "new-round"
    | "room-locked"
    | "developer-skipped"
    | "room-opened";
  data?: any;
};

export const useWebsocketStore = defineStore("websocket", (): WebsocketStore => {
  const username = ref("");
  const userRole: Ref<Role> = ref(Role.Empty);
  const userRoomId = ref("");
  const websocket: Ref<WebSocket | null> = ref(null);
  const usersInRoom: Ref<UserOverview> = ref({
    developerList: [],
    productOwnerList: [],
  });
  const roundState: Ref<RoundState> = ref(RoundState.Waiting);
  const ticketToGuess = ref("");
  const guess = ref(0);
  const didSkip = ref(false);
  const showAllGuesses = ref(false);
  const possibleGuesses: Ref<PossibleGuess[]> = ref([]);
  const permissions: Ref<Permissions> = ref({ room: { canLock: false } });
  const roomIsLocked: Ref<boolean> = ref(false);
  const developerDone: Ref<DeveloperDone[]> = ref([]);

  const isConnected = computed(() => websocket.value !== null);

  function disconnect() {
    websocket.value?.close();
    websocket.value = null;
    permissions.value = { room: { canLock: false } };
  }

  async function connect(name: string, role: Role, roomId: string): Promise<void> {
    username.value = name;
    userRole.value = role;
    userRoomId.value = roomId;

    const roleUrl = role === Role.Developer ? "developer" : "product-owner";
    let wsUrl = `wss://${window.location.host}/api/estimation/room/${roomId}/${roleUrl}?name=${name}`;
    if (window.location.protocol !== "https:") {
      wsUrl = `ws://${window.location.host}/api/estimation/room/${roomId}/${roleUrl}?name=${name}`;
    }
    websocket.value = new WebSocket(wsUrl);
    await waitForOpenConnection(websocket.value);

    websocket.value!.onerror = () => {
      websocket.value!.close();
    };

    websocket.value!.onclose = () => {
      websocket.value?.close();
    };

    websocket.value!.onmessage = async (message: MessageEvent) => {
      const decoded = JSON.parse(message.data) as ReceivableWebsocketMessage;
      switch (decoded.type) {
        case "leave":
        case "join":
        case "developer-skipped":
        case "developer-guessed":
          await fetchUsersInRoom();
          break;
        case "estimate":
          roundState.value = RoundState.InProgress;
          ticketToGuess.value = decoded.data;
          break;
        case "you-guessed":
          guess.value = decoded.data;
          didSkip.value = false;
          break;
        case "you-skipped":
          guess.value = 0;
          didSkip.value = true;
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
        case "room-opened":
          await fetchRoomIsLocked();
          break;
        case "new-round":
          resetRound();
          await fetchUsersInRoom();
          break;
      }
    };
  }

  function waitForOpenConnection(socket: WebSocket): Promise<boolean> {
    return new Promise((resolve) => {
      if (socket.readyState !== socket.OPEN) {
        socket.addEventListener("open", () => {
          resolve(true);
        });
      } else {
        resolve(true);
      }
    });
  }

  function send(message: SendableWebsocketMessage) {
    if (!websocket.value) {
      throw new Error("Can not send message without a connection");
    }

    websocket.value?.send(JSON.stringify(message));
  }

  function resetRound() {
    ticketToGuess.value = "";
    guess.value = 0;
    roundState.value = RoundState.Waiting;
    showAllGuesses.value = false;
    didSkip.value = false;
    developerDone.value = [];
  }

  async function userExistsInRoom(name: string, roomId: string): Promise<boolean> {
    const response = await fetch(`/api/estimation/room/${roomId}/users/exists?name=${name}`);
    return ((await response.json()) as { exists: boolean }).exists;
  }

  async function isRoundInRoomInProgress(roomId: string): Promise<boolean> {
    const response = await fetch(`/api/estimation/room/${roomId}/state`);
    return ((await response.json()) as { inProgress: boolean }).inProgress;
  }

  async function isRoomLocked(roomId: string): Promise<boolean> {
    const response = await fetch(`/api/estimation/room/${roomId}/state`);
    return ((await response.json()) as { isLocked: boolean }).isLocked;
  }

  async function passwordMatchesRoom(roomId: string, password: string): Promise<boolean> {
    const response = await fetch(`/api/estimation/room/${roomId}/authenticate`, {
      method: "POST",
      body: JSON.stringify({ password }),
    });

    if (!response.ok) {
      return false;
    }

    return ((await response.json()) as { ok: boolean }).ok;
  }

  async function fetchUsersInRoom() {
    const response = await fetch(`/api/estimation/room/${userRoomId.value}/users`);
    if (!response.ok) {
      usersInRoom.value = {
        productOwnerList: [],
        developerList: [],
      };
      return;
    }

    usersInRoom.value = await response.json();
  }

  async function fetchActiveRooms(): Promise<string[]> {
    return (await fetch("/api/estimation/room/rooms")).json();
  }

  async function fetchPossibleGuesses() {
    const response = await fetch("/api/estimation/possible-guesses");
    if (!response.ok) {
      possibleGuesses.value = [];
      return;
    }

    possibleGuesses.value = await response.json();
  }

  async function fetchPermissions(): Promise<void> {
    const response = await fetch(
      `/api/estimation/room/${userRoomId.value}/${username.value}/permissions`,
    );
    if (!response.ok) {
      permissions.value = {
        room: {
          canLock: false,
        },
      };
      return;
    }
    permissions.value = (await response.json()).permissions;
    return;
  }

  async function fetchRoomIsLocked(): Promise<boolean> {
    const response = await fetch(`/api/estimation/room/${userRoomId.value}/state`);
    if (!response.ok) {
      roomIsLocked.value = false;
      return roomIsLocked.value;
    }

    roomIsLocked.value = ((await response.json()) as { isLocked: boolean }).isLocked;
    return roomIsLocked.value;
  }

  return {
    connect,
    disconnect,
    isConnected,
    usersInRoom,
    isRoundInRoomInProgress,
    isRoomLocked,
    roomId: userRoomId,
    possibleGuesses,
    username,
    userExistsInRoom,
    userRole,
    roundState,
    send,
    ticketToGuess,
    guess,
    didSkip,
    resetRound,
    showAllGuesses,
    fetchActiveRooms,
    fetchPossibleGuesses,
    fetchPermissions,
    fetchRoomIsLocked,
    passwordMatchesRoom,
    permissions,
    roomIsLocked,
    developerDone,
  };
});
