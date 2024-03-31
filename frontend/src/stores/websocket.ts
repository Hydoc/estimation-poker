import type { Ref } from "vue";
import { computed, ref } from "vue";
import { defineStore } from "pinia";
import { type PossibleGuess, type UserOverview, type Permissions } from "@/components/types";
import { Role, RoundState } from "@/components/types";

type WebsocketStore = {
  connect(name: string, role: string, roomId: string): void;
  disconnect(): void;
  resetRound(): void;
  userExistsInRoom(name: string, roomId: string): Promise<boolean>;
  fetchActiveRooms(): Promise<string[]>;
  send(message: SendableWebsocketMessage): void;
  isRoundInRoomInProgress(roomId: string): Promise<boolean>;
  fetchPossibleGuesses(): Promise<void>;
  fetchPermissions(): Promise<Permissions>;
  fetchRoomIsLocked(): Promise<boolean>;
  username: Ref<string>;
  isConnected: Ref<boolean>;
  usersInRoom: Ref<UserOverview>;
  roomId: Ref<string>;
  userRole: Ref<Role>;
  roundState: Ref<RoundState>;
  ticketToGuess: Ref<string>;
  guess: Ref<number>;
  showAllGuesses: Ref<boolean>;
  possibleGuesses: Ref<PossibleGuess[]>;
  permissions: Ref<Permissions>;
  roomIsLocked: Ref<boolean>;
};

export type SendableWebsocketMessageType = "estimate" | "guess" | "reveal" | "new-round" | "lock-room";

type SendableWebsocketMessage = {
  type: SendableWebsocketMessageType;
  data?: any;
};

type ReceivableWebsocketMessage = {
  type:
    | "join"
    | "leave"
    | "estimate"
    | "developer-guessed"
    | "everyone-guessed"
    | "you-guessed"
    | "reveal"
    | "reset-round"
    | "room-locked";
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
  const showAllGuesses = ref(false);
  const possibleGuesses: Ref<PossibleGuess[]> = ref([]);
  const permissions: Ref<Permissions> = ref({ room: { canLock: false } });
  const roomIsLocked: Ref<boolean> = ref(false);

  const isConnected = computed(() => websocket.value !== null);

  function disconnect() {
    websocket.value?.close();
    websocket.value = null;
    permissions.value = { room: { canLock: false } };
  }

  function connect(name: string, role: Role, roomId: string): void {
    username.value = name;
    userRole.value = role;
    userRoomId.value = roomId;

    const roleUrl = role === Role.Developer ? "developer" : "product-owner";
    let wsUrl = `wss://${window.location.host}/api/estimation/room/${roomId}/${roleUrl}?name=${name}`
    if (window.location.protocol !== "https:") {
      wsUrl = `ws://${window.location.host}/api/estimation/room/${roomId}/${roleUrl}?name=${name}`
    }
    websocket.value = new WebSocket(wsUrl);

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
        case "developer-guessed":
          await fetchUsersInRoom();
          break;
        case "estimate":
          roundState.value = RoundState.InProgress;
          ticketToGuess.value = decoded.data;
          break;
        case "you-guessed":
          guess.value = decoded.data;
          break;
        case "everyone-guessed":
          await fetchUsersInRoom();
          roundState.value = RoundState.End;
          break;
        case "reveal":
          showAllGuesses.value = true;
          break;
        case "room-locked":
          await fetchRoomIsLocked();
          break;
        case "reset-round":
          resetRound();
          await fetchUsersInRoom();
          break;
      }
    };
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
  }

  async function userExistsInRoom(name: string, roomId: string): Promise<boolean> {
    const response = await fetch(`/api/estimation/room/${roomId}/users/exists?name=${name}`);
    return ((await response.json()) as { exists: boolean }).exists;
  }

  async function isRoundInRoomInProgress(roomId: string): Promise<boolean> {
    const response = await fetch(`/api/estimation/room/${roomId}/state`);
    return ((await response.json()) as { inProgress: boolean }).inProgress;
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
  
  async function fetchPermissions(): Promise<Permissions> {
    const response = await fetch(`/api/estimation/room/${userRoomId.value}/${username.value}/permissions`);
    if (!response.ok) {
      permissions.value = {
        room: {
          canLock: false,
        }
      };
      return permissions.value;
    }
    permissions.value = (await response.json()).permissions;
    return permissions.value;
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
    roomId: userRoomId,
    possibleGuesses,
    username,
    userExistsInRoom,
    userRole,
    roundState,
    send,
    ticketToGuess,
    guess,
    resetRound,
    showAllGuesses,
    fetchActiveRooms,
    fetchPossibleGuesses,
    fetchPermissions,
    fetchRoomIsLocked,
    permissions,
    roomIsLocked,
  };
});
