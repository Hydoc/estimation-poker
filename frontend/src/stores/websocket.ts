import type {Ref} from "vue";
import {computed, ref} from "vue";
import {defineStore} from "pinia";
import type {UserOverview} from "@/components/types";
import {Role, RoundState} from "@/components/types";

type WebsocketStore = {
  connect(name: string, role: string, roomId: string): void;
  disconnect(): void;
  resetRound(): void;
  userExistsInRoom(name: string, roomId: string): Promise<boolean>;
  send(message: SendableWebsocketMessage): void;
  username: Ref<string>;
  isConnected: Ref<boolean>;
  usersInRoom: Ref<UserOverview>;
  roomId: Ref<string>;
  userRole: Ref<Role>;
  roundState: Ref<RoundState>;
  ticketToGuess: Ref<string>;
  guess: Ref<number>;
};

export type SendableWebsocketMessageType = "estimate" | "guess" | "reveal";

type SendableWebsocketMessage = {
  type: SendableWebsocketMessageType;
  data?: any;
};

type ReceivableWebsocketMessage = {
  type: "join" | "leave" | "estimate" | "developer-guessed" | "everyone-guessed" | "you-guessed";
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

  const isConnected = computed(() => websocket.value !== null);

  function disconnect() {
    websocket.value?.close();
    websocket.value = null;
  }

  function connect(name: string, role: Role, roomId: string): void {
    username.value = name;
    userRole.value = role;
    userRoomId.value = roomId;

    const roleUrl = role === Role.Developer ? "developer" : "product-owner";
    websocket.value = new WebSocket(`ws://localhost:8090/room/${roomId}/${roleUrl}?name=${name}`);

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
      }
      console.log("Message flew", message);
    };
  }

  function send(message: SendableWebsocketMessage) {
    if (!websocket.value) {
      throw new Error("Can not estimate ticket without a connection");
    }

    websocket.value?.send(JSON.stringify(message));
  }

  function resetRound() {
    ticketToGuess.value = "";
    guess.value = 0;
    roundState.value = RoundState.Waiting;
  }

  async function userExistsInRoom(name: string, roomId: string): Promise<boolean> {
    const response = await fetch(`http://localhost:8090/room/${roomId}/users/exists?name=${name}`);
    return (await response.json()).exists;
  }

  async function fetchUsersInRoom() {
    const response = await fetch(`http://localhost:8090/room/${userRoomId.value}/users`);
    if (!response.ok) {
      usersInRoom.value = {
        productOwnerList: [],
        developerList: [],
      };
      return;
    }

    usersInRoom.value = await response.json();
  }

  return {
    connect,
    disconnect,
    isConnected,
    usersInRoom,
    roomId: userRoomId,
    username,
    userExistsInRoom,
    userRole,
    roundState,
    send,
    ticketToGuess,
    guess,
    resetRound,
  };
});
